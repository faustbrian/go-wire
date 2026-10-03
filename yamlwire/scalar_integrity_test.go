package yamlwire_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/yamlwire"
	"go.yaml.in/yaml/v4"
)

func TestEncodePreservesScalarSyntaxAndBlockContents(t *testing.T) {
	for _, scenario := range []struct {
		name  string
		value any
	}{
		{"strip-like suffix", map[string]string{"text": "000- >-"}}, {"keep-like suffix", []string{"000- |+"}}, {"literal suffix", map[string][]string{"text": {"000- |"}}}, {"quoted scalar", map[string]string{"text": "prefix: >"}}, {"quoted key", map[string]string{"colon: key": "\t\n0"}}, {"plain apostrophe key", map[string]string{"don't": "\t\n0"}}, {"root clip", "\t\n0\n"}, {"sequence keep", []string{"\t\n0\n\n"}}, {"plain root", "000- >"}, {"plain map", map[string]string{"text": "000- >"}},
		{"multiline key content", map[string]string{"first\nfoo: |\n- >\nlast": "value"}},
		{"multiline tab key", map[string]string{"\t\n0": "value"}},
		{"nested explicit key", []map[string]string{{"first\nfoo: |\n- >\nlast": "\t\n0", "\t\n0": "value"}}},
		{"plain sequence", []string{"000- >"}},
		{"nested sequence", [][]string{{"000- >", "\t\n0"}}},
		{"nested map", []map[string]string{{"text": "\t\n0", "sibling": "normal\nnext"}}},
		{"block content", map[string]string{"text": "first\n000- >\nfoo: |\n- >\nlast"}},
		{"real map block", map[string]string{"text": "\t\n0"}},
		{"real root block", "\t\n0"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			for _, options := range []yamlwire.EncodeOptions{{}, {Indent: 2}, {Indent: 4}, {Indent: 4, DefaultSequenceIndent: true}, {Indent: 9}} {
				payload, err := yamlwire.Encode(scenario.value, options)
				if err != nil {
					t.Fatal(err)
				}
				got := reflect.New(reflect.TypeOf(scenario.value))
				err = yamlwire.Decode(payload, got.Interface(), yamlwire.DecodeOptions{})
				t.Logf("options=%+v payload=%q", options, payload)
				independent := reflect.New(reflect.TypeOf(scenario.value))
				if err := yaml.Load(payload, independent.Interface(), yaml.WithV4Defaults()); err != nil || !reflect.DeepEqual(independent.Elem().Interface(), scenario.value) {
					t.Errorf("independent round trip = %#v error=%v, want %#v", independent.Elem().Interface(), err, scenario.value)
				}
				if err != nil || !reflect.DeepEqual(got.Elem().Interface(), scenario.value) {
					t.Errorf("round trip = %#v error=%v, want %#v", got.Elem().Interface(), err, scenario.value)
				}
			}
		})
	}
}

func TestEncodeWriterPreservesScalarSyntaxAndFinalQuota(t *testing.T) {
	t.Parallel()
	for _, value := range []any{
		struct {
			Text   string `yaml:"text"`
			Number int64  `yaml:"number"`
			Flag   bool   `yaml:"flag"`
		}{"000- >", -56, true},
		map[string]string{"text": "\t\n0"},
	} {
		payload, err := yamlwire.Encode(value, yamlwire.EncodeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got := reflect.New(reflect.TypeOf(value))
		if err := yamlwire.Decode(payload, got.Interface(), yamlwire.DecodeOptions{}); err != nil || !reflect.DeepEqual(got.Elem().Interface(), value) {
			t.Fatalf("round trip = %#v, error=%v, want %#v", got.Elem().Interface(), err, value)
		}
		for _, allowance := range []int64{-1, 0, 1} {
			options := yamlwire.EncodeOptions{MaxBytes: int64(len(payload)) + allowance}
			limited, err := yamlwire.Encode(value, options)
			var writer bytes.Buffer
			writer.WriteString("prior")
			writeErr := yamlwire.EncodeWriter(&writer, value, options)
			if allowance < 0 {
				if limited != nil || !errors.Is(err, wire.ErrSizeLimit) || !errors.Is(err, yamlwire.ErrPayloadTooLarge) {
					t.Fatalf("below final quota returned %q, %v", limited, err)
				}
				if !errors.Is(writeErr, wire.ErrSizeLimit) || !errors.Is(writeErr, yamlwire.ErrPayloadTooLarge) || writer.String() != "prior" {
					t.Fatalf("failed writer published %q, %v", writer.String(), writeErr)
				}
			} else if err != nil || writeErr != nil || !bytes.Equal(limited, payload) || writer.String() != "prior"+string(payload) {
				t.Fatalf("successful quota returned %q, %v; writer %q, %v", limited, err, writer.String(), writeErr)
			}
		}
	}
}

func TestEncodePreservesSupportedNodeScalarSemantics(t *testing.T) {
	t.Parallel()
	for _, scalar := range []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!vendor", Value: "000- >", Style: yaml.TaggedStyle},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "first\nfoo: |\n- >\nlast", Style: yaml.SingleQuotedStyle},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "first\n- |\nlast\n", Style: yaml.LiteralStyle, LineComment: "block"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "first\n- |\nlast\n", Style: yaml.FoldedStyle, Anchor: "content"},
	} {
		wantValue, wantTag := scalar.Value, scalar.Tag
		node := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "text"}, scalar,
		}}
		for _, value := range []any{node, scalarNodeMarshaler{node}} {
			payload, err := yamlwire.Encode(value, yamlwire.EncodeOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var got yaml.Node
			if err := yaml.Load(payload, &got, yaml.WithV4Defaults()); err != nil {
				t.Fatalf("independent parse of %q failed: %v", payload, err)
			}
			if len(got.Content) != 1 || len(got.Content[0].Content) != 2 {
				t.Fatalf("node collection changed: %#v", got)
			}
			decoded := got.Content[0].Content[1]
			if decoded.Value != wantValue || decoded.Tag != wantTag {
				t.Fatalf("node scalar = %q/%q, want %q/%q; output %q", decoded.Value, decoded.Tag, wantValue, wantTag, payload)
			}
			repeated, err := yamlwire.Encode(value, yamlwire.EncodeOptions{})
			if err != nil || !bytes.Equal(repeated, payload) {
				t.Fatalf("nondeterministic encode: %q, %v", repeated, err)
			}
			var output bytes.Buffer
			if err := yamlwire.EncodeWriter(&output, value, yamlwire.EncodeOptions{}); err != nil || !bytes.Equal(output.Bytes(), payload) {
				t.Fatalf("writer output %q, %v", output.Bytes(), err)
			}
		}
	}
}

type scalarNodeMarshaler struct{ node *yaml.Node }

func (v scalarNodeMarshaler) MarshalYAML() (any, error) { return v.node, nil }

func TestEncodePreservesQuotedContinuationContexts(t *testing.T) {
	text := "first\nfoo: |\n- >\nlast's end"
	quoted := func() *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: text, Style: yaml.SingleQuotedStyle}
	}
	plain := func(value string) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
	}
	for _, scenario := range []struct {
		name string
		node *yaml.Node
	}{
		{"root", quoted()},
		{"sequence", &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{quoted(), plain("sibling")}}},
		{"explicit quoted key", &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{quoted(), plain("sibling")}}},
		{"flow sequence", &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{quoted(), plain("sibling")}}},
		{"flow mapping", &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle, Content: []*yaml.Node{plain("text"), quoted(), plain("sibling"), plain("unchanged")}}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			want := scalarNodeSemantics(scenario.node)
			for _, options := range []yamlwire.EncodeOptions{{}, {Indent: 4}, {Indent: 9}} {
				payload, err := yamlwire.Encode(scenario.node, options)
				if err != nil {
					t.Fatal(err)
				}
				var got yaml.Node
				if err := yaml.Load(payload, &got, yaml.WithV4Defaults()); err != nil {
					t.Fatalf("parse %q: %v", payload, err)
				}
				if len(got.Content) != 1 || !reflect.DeepEqual(scalarNodeSemantics(got.Content[0]), want) {
					t.Fatalf("node semantics changed: %q", payload)
				}
				var output bytes.Buffer
				if err := yamlwire.EncodeWriter(&output, scenario.node, options); err != nil || !bytes.Equal(output.Bytes(), payload) {
					t.Fatalf("writer output %q error=%v", output.Bytes(), err)
				}
			}
		})
	}
}

type nodeSemantics struct {
	kind     yaml.Kind
	value    string
	children []nodeSemantics
}

func scalarNodeSemantics(node *yaml.Node) nodeSemantics {
	result := nodeSemantics{kind: node.Kind, value: node.Value}
	for _, child := range node.Content {
		result.children = append(result.children, scalarNodeSemantics(child))
	}
	return result
}
