package yamlwire_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire/v2"
	"github.com/faustbrian/go-wire/v2/yamlwire"
	"go.yaml.in/yaml/v4"
)

func TestEncodePreservesScalarSyntaxAndBlockContents(t *testing.T) {
	for _, scenario := range []struct {
		name  string
		value any
	}{
		{"strip-like suffix", map[string]string{"text": "000- >-"}}, {"keep-like suffix", []string{"000- |+"}}, {"literal suffix", map[string][]string{"text": {"000- |"}}}, {"quoted scalar", map[string]string{"text": "prefix: >"}}, {"quoted key", map[string]string{"colon: key": "\t\n0"}}, {"plain apostrophe key", map[string]string{"don't": "\t\n0"}}, {"root clip", "\t\n0\n"}, {"sequence keep", []string{"\t\n0\n\n"}}, {"plain root", "000- >"}, {"plain map", map[string]string{"text": "000- >"}},
		{"plain block-like wrap suffix", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |"},
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
		"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |",
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
		{"inline flow sequence", &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{plain("foo: | #body")}}},
		{"inline flow mapping key", &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle, Content: []*yaml.Node{plain("foo: | #key"), plain("unchanged")}}},
		{"inline flow mapping value", &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle, Content: []*yaml.Node{plain("text"), plain("foo: > #value")}}},
		{"nested inline flow", &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{plain("foo: |- #nested")}}}}},
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

func TestEncodePreservesExplicitIndentAndEscapedKeys(t *testing.T) {
	scalar := func(value string, style yaml.Style) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: style}
	}
	mapping := func(key, value *yaml.Node) *yaml.Node {
		return &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{key, value}}
	}
	for _, scenario := range []struct {
		name string
		node *yaml.Node
	}{
		{"leading spaces", scalar(" leading\n last\n", yaml.LiteralStyle)},
		{"empty flow sequence key", mapping(&yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}, scalar("\t\n0", yaml.LiteralStyle))},
		{"empty flow mapping key", mapping(&yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle}, scalar("\t\n0", yaml.LiteralStyle))},
		{"double quoted key", mapping(scalar("quoted\"\\key: label", yaml.DoubleQuotedStyle), scalar(" leading\n last", yaml.LiteralStyle))},
		{"single quoted key", mapping(scalar("don't: label", yaml.SingleQuotedStyle), scalar(" leading\n last", yaml.LiteralStyle))},
		{"double quoted scalar", scalar("escaped \\ \"\nfoo: |\n- >", yaml.DoubleQuotedStyle)},
		{"compact sequence mapping", &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{mapping(scalar("quoted\"\\key: label", yaml.DoubleQuotedStyle), scalar(" leading\n last", yaml.LiteralStyle))}}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			want := scalarNodeSemantics(scenario.node)
			for _, indent := range []int{0, 2, 4, 9} {
				for _, sequenceIndent := range []bool{false, true} {
					options := yamlwire.EncodeOptions{Indent: indent, DefaultSequenceIndent: sequenceIndent}
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
					output.Reset()
					output.WriteString("prior")
					options.MaxBytes = int64(len(payload) - 1)
					if err := yamlwire.EncodeWriter(&output, scenario.node, options); !errors.Is(err, wire.ErrSizeLimit) || output.String() != "prior" {
						t.Fatalf("limited writer published %q error=%v", output.Bytes(), err)
					}
				}
			}
		})
	}
}

func TestEncodePreservesMixedBlockContinuationBoundaries(t *testing.T) {
	t.Parallel()
	scalar := func(value string, style yaml.Style) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: style}
	}
	for _, indent := range []int{0, 2, 3, 4, 5, 6, 7, 8, 9} {
		for _, sequenceIndent := range []bool{false, true} {
			// Fresh nodes avoid the pinned dumper's borrowed style/tag updates
			// carrying state between independently configured encodings.
			makeNode := func() *yaml.Node {
				anchored := scalar(" leading\n foo: |\n - >\n last\n", yaml.TaggedStyle|yaml.LiteralStyle)
				anchored.Tag, anchored.Anchor = "!vendor", "block"
				return &yaml.Node{Kind: yaml.MappingNode, Tag: "!collection", Anchor: "document", Style: yaml.TaggedStyle, Content: []*yaml.Node{
					scalar("already explicit", 0), anchored,
					scalar("alias", 0), {Kind: yaml.AliasNode, Value: "block", Alias: anchored},
					scalar("quoted continuation", 0), scalar("first\nfoo: |\n- >\nlast's end", yaml.SingleQuotedStyle),
					scalar("flow continuation", 0), {Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{
						scalar("first\nfoo: |\n- >\nlast's end", yaml.SingleQuotedStyle), scalar("plain sibling", 0),
						scalar("earlier's quote\nfoo: |\n- >\nlast", yaml.SingleQuotedStyle),
					}},
					scalar("protected body", 0), scalar("first\n['unclosed \"\nfoo: |\n- >\nlast\n", yaml.LiteralStyle),
					scalar("plain punctuation", 0), scalar("normal text, 'not a quote", 0),
					scalar("later block", 0), {Kind: yaml.ScalarNode, Tag: "!!str", Value: "\t\n0", Style: yaml.LiteralStyle, HeadComment: "note: 'unclosed [", LineComment: "label: |", FootComment: "note: \"unclosed {"},
					scalar("folded strip", 0), scalar(" leading\n last", yaml.FoldedStyle),
					scalar("folded keep", 0), scalar(" leading\n last\n\n", yaml.FoldedStyle),
					scalar("nested blocks", 0), {Kind: yaml.SequenceNode, Content: []*yaml.Node{
						{Kind: yaml.MappingNode, Content: []*yaml.Node{
							scalar("don't: label", yaml.SingleQuotedStyle), scalar("\t\n0", yaml.LiteralStyle),
							scalar("last", 0), scalar(" leading\n final", yaml.LiteralStyle),
						}},
					}},
					{Kind: yaml.SequenceNode, Content: []*yaml.Node{scalar("collection", 0), scalar("key", 0)}}, scalar("\t\n0", yaml.LiteralStyle),
					{Kind: yaml.SequenceNode, Content: []*yaml.Node{scalar("\t\n0", yaml.LiteralStyle), scalar("block key sibling", 0)}}, {Kind: yaml.SequenceNode, Content: []*yaml.Node{scalar("\t\n1", yaml.LiteralStyle), scalar("block value sibling", 0)}},
					scalar("nested flow block", 0), {Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{scalar("first", 0), scalar("foo: |\n- >", yaml.LiteralStyle)}},
					scalar("inline flow sequence", 0), {Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{scalar("foo: | #body", yaml.SingleQuotedStyle), scalar("foo: >- #folded", yaml.DoubleQuotedStyle)}},
					scalar("inline flow mapping", 0), {Kind: yaml.MappingNode, Style: yaml.FlowStyle, Content: []*yaml.Node{scalar("foo: | #key", yaml.SingleQuotedStyle), scalar("foo: > #value", yaml.SingleQuotedStyle)}},
					scalar("nested inline flow", 0), {Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{{Kind: yaml.MappingNode, Style: yaml.FlowStyle, Content: []*yaml.Node{scalar("foo: | #nested", yaml.SingleQuotedStyle), scalar("unchanged", 0)}}}},
					{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{scalar("flow", 0), scalar("key", 0)}}, scalar("\t\n0", yaml.LiteralStyle),
					scalar("final sibling", 0), scalar("unchanged", 0),
				}}
			}
			options := yamlwire.EncodeOptions{Indent: indent, DefaultSequenceIndent: sequenceIndent}
			want := scalarNodeSemantics(makeNode())
			payload, err := yamlwire.Encode(makeNode(), options)
			if err != nil {
				t.Fatalf("options=%+v: %v", options, err)
			}
			for _, decode := range []func([]byte, *yaml.Node) error{
				func(payload []byte, got *yaml.Node) error { return yaml.Load(payload, got, yaml.WithV4Defaults()) },
				func(payload []byte, got *yaml.Node) error {
					return yamlwire.Decode(payload, got, yamlwire.DecodeOptions{})
				},
			} {
				var got yaml.Node
				if err := decode(payload, &got); err != nil || len(got.Content) != 1 || !reflect.DeepEqual(scalarNodeSemantics(got.Content[0]), want) {
					t.Fatalf("options=%+v: mixed document changed: %q, error=%v", options, payload, err)
				}
				mapping := got.Content[0]
				anchored, alias := mapping.Content[1], mapping.Content[3]
				if mapping.Tag != "!collection" || mapping.Anchor != "document" || anchored.Tag != "!vendor" || anchored.Anchor != "block" || alias.Alias != anchored {
					t.Fatalf("options=%+v: scalar tag, anchor, or alias target changed: %q", options, payload)
				}
			}
			repeated, err := yamlwire.Encode(makeNode(), options)
			if err != nil || !bytes.Equal(repeated, payload) {
				t.Fatalf("options=%+v: nondeterministic document: %q, error=%v", options, repeated, err)
			}
			for _, allowance := range []int64{-1, 0, 1} {
				limited := options
				limited.MaxBytes = int64(len(payload)) + allowance
				encoded, encodeErr := yamlwire.Encode(makeNode(), limited)
				var output bytes.Buffer
				output.WriteString("prior")
				writeErr := yamlwire.EncodeWriter(&output, makeNode(), limited)
				if allowance < 0 {
					if encoded != nil || !errors.Is(encodeErr, wire.ErrSizeLimit) || !errors.Is(encodeErr, yamlwire.ErrPayloadTooLarge) || !errors.Is(writeErr, wire.ErrSizeLimit) || output.String() != "prior" {
						t.Fatalf("options=%+v: final quota published %q/%q, errors=%v/%v", limited, encoded, output.String(), encodeErr, writeErr)
					}
				} else if encodeErr != nil || writeErr != nil || !bytes.Equal(encoded, payload) || output.String() != "prior"+string(payload) {
					t.Fatalf("options=%+v: exact final quota changed output, errors=%v/%v", limited, encodeErr, writeErr)
				}
			}
		}
	}
}

func TestEncodePreservesNonBlockTrailingComments(t *testing.T) {
	for _, sequence := range []bool{false, true} {
		makeNode := func() *yaml.Node {
			scalar := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "unchanged", LineComment: "foo: |"}
			if !sequence {
				return scalar
			}
			return &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{
				scalar,
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "\t\n0", Style: yaml.LiteralStyle},
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "sibling"},
			}}
		}
		for _, indent := range []int{0, 4, 9} {
			options := yamlwire.EncodeOptions{Indent: indent}
			payload, err := yamlwire.Encode(makeNode(), options)
			if err != nil {
				t.Fatal(err)
			}
			var got yaml.Node
			if err := yaml.Load(payload, &got, yaml.WithV4Defaults()); err != nil || len(got.Content) != 1 {
				t.Fatalf("parse %q: %v", payload, err)
			}
			node := got.Content[0]
			if !reflect.DeepEqual(scalarNodeSemantics(node), scalarNodeSemantics(makeNode())) {
				t.Fatalf("authored scalar content changed: %q", payload)
			}
			if sequence {
				node = node.Content[0]
			}
			if node.LineComment != "# foo: |" {
				t.Fatalf("authored trailing comment changed to %q: %q", node.LineComment, payload)
			}
			for _, allowance := range []int64{-1, 0, 1} {
				limited := options
				limited.MaxBytes = int64(len(payload)) + allowance
				encoded, encodeErr := yamlwire.Encode(makeNode(), limited)
				var output bytes.Buffer
				output.WriteString("prior")
				writeErr := yamlwire.EncodeWriter(&output, makeNode(), limited)
				if allowance < 0 {
					if encoded != nil || !errors.Is(encodeErr, wire.ErrSizeLimit) || !errors.Is(writeErr, wire.ErrSizeLimit) || output.String() != "prior" {
						t.Fatalf("below quota published %q/%q: %v/%v", encoded, output.String(), encodeErr, writeErr)
					}
				} else if encodeErr != nil || writeErr != nil || !bytes.Equal(encoded, payload) || output.String() != "prior"+string(payload) {
					t.Fatalf("admitted quota changed output: %v/%v", encodeErr, writeErr)
				}
			}
		}
	}
}
