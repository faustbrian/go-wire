package jsonwire_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/jsonwire"
)

func TestDecodeCanReplaceInvalidUTF8LikeEncodingJSON(t *testing.T) {
	t.Parallel()

	// Each invalid byte inside a string becomes one U+FFFD, as encoding/json does.
	payload := []byte("{\"value\":\"p\xe4iv\xe4\xe4\",\"n\xffme\":1}")
	var target map[string]any
	if err := jsonwire.Decode(payload, &target, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true}); err != nil {
		t.Fatalf("Decode(ReplaceInvalidUTF8) error = %v", err)
	}
	if target["value"] != "p�iv��" || target["n�me"] != float64(1) {
		t.Fatalf("Decode(ReplaceInvalidUTF8) = %#v", target)
	}

	var strict map[string]any
	assertKind(t, jsonwire.Decode(payload, &strict, jsonwire.DecodeOptions{}), wire.ErrParse)

	// Invalid bytes outside a string remain malformed JSON.
	var outside any
	assertKind(t, jsonwire.Decode([]byte("[1,\xff]"), &outside, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true}), wire.ErrParse)
}

func TestDuplicateNamesCompareReplacedInvalidUTF8(t *testing.T) {
	t.Parallel()

	// Both names decode to "a�", so they are the same member name.
	payload := []byte("{\"a\xff\":1,\"a\xfe\":2}")
	var target map[string]int
	assertKind(t, jsonwire.Decode(payload, &target, jsonwire.DecodeOptions{
		ReplaceInvalidUTF8: true, DisallowDuplicateNames: true,
	}), wire.ErrValidation)
}

func TestDecodeReaderCanReplaceInvalidUTF8(t *testing.T) {
	t.Parallel()

	var target []string
	err := jsonwire.DecodeReader(strings.NewReader("[\"\xc3\"]"), &target, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true})
	if err != nil || len(target) != 1 || target[0] != "�" {
		t.Fatalf("DecodeReader(ReplaceInvalidUTF8) = %#v, %v", target, err)
	}
}

func TestDecodeCanBoundNestedLevels(t *testing.T) {
	t.Parallel()

	nested := func(levels int) string {
		return strings.Repeat(`{"a":[`, levels/2) + strings.Repeat(`[`, levels%2) + `1` +
			strings.Repeat(`]`, levels%2) + strings.Repeat(`]}`, levels/2)
	}
	for _, testCase := range []struct {
		payload   string
		maximum   int
		duplicate bool
		ok        bool
	}{
		{`1`, 1, false, true},
		{`"text"`, 1, true, true},
		{`[]`, 1, false, true},
		{`{}`, 1, true, true},
		{nested(4), 4, false, true},
		{nested(4), 4, true, true},
		{nested(5), 4, false, false},
		{nested(5), 4, true, false},
		{nested(3), 4, false, true},
		{`[[1],[2],[[3]]]`, 2, false, false},
		{`[[1],[2],[3]]`, 2, false, true},
		{strings.Repeat("[", 10_000) + strings.Repeat("]", 10_000), 64, false, false},
		{nested(100), 0, false, true},
	} {
		var target any
		err := jsonwire.Decode([]byte(testCase.payload), &target, jsonwire.DecodeOptions{
			MaxNestedLevels: testCase.maximum, DisallowDuplicateNames: testCase.duplicate,
		})
		if testCase.ok {
			if err != nil {
				t.Errorf("Decode(%d-byte payload, max %d) error = %v", len(testCase.payload), testCase.maximum, err)
			}
			continue
		}
		assertKind(t, err, wire.ErrSizeLimit)
		if target != nil {
			t.Errorf("Decode(over limit) mutated target = %#v", target)
		}
	}
}

func TestDecodeRejectsNestedLevelsOutsideEncodingJSONRange(t *testing.T) {
	t.Parallel()

	var target any
	assertKind(t, jsonwire.Decode([]byte(`1`), &target, jsonwire.DecodeOptions{MaxNestedLevels: -1}), wire.ErrValidation)
	assertKind(t, jsonwire.Decode([]byte(`1`), &target, jsonwire.DecodeOptions{MaxNestedLevels: 10_001}), wire.ErrValidation)
	deepest := strings.Repeat("[", 10_000) + strings.Repeat("]", 10_000)
	if err := jsonwire.Decode([]byte(deepest), &target, jsonwire.DecodeOptions{MaxNestedLevels: 10_000}); err != nil {
		t.Fatalf("Decode(10,000 levels, max 10,000) error = %v", err)
	}
}

func TestStructurePreflightAcceptsNumbersOutsideFloat64Range(t *testing.T) {
	t.Parallel()

	for name, options := range map[string]jsonwire.DecodeOptions{
		"duplicate names": {DisallowDuplicateNames: true},
		"nested levels":   {MaxNestedLevels: 4},
	} {
		var target struct {
			Value json.Number `json:"value"`
		}
		if err := jsonwire.Decode([]byte(`{"value":1e400}`), &target, options); err != nil || target.Value != "1e400" {
			t.Fatalf("%s: Decode(1e400) = %q, %v", name, target.Value, err)
		}
	}
}

func TestNestedLevelsOnlyStillRejectsTrailingValues(t *testing.T) {
	t.Parallel()

	var target any
	assertKind(t, jsonwire.Decode([]byte(`[1] [2]`), &target, jsonwire.DecodeOptions{MaxNestedLevels: 4}), wire.ErrParse)
}

// encoding/json is the reference for replacement: every shape must decode to
// the same value through jsonwire with ReplaceInvalidUTF8 and json.Unmarshal.
func TestReplaceInvalidUTF8MatchesEncodingJSON(t *testing.T) {
	t.Parallel()

	for name, payload := range map[string]string{
		"truncated three-byte sequence": "{\"v\":\"a\xe2\x82\"}",
		"truncated four-byte sequence":  "{\"v\":\"a\xf0\x9f\x98\"}",
		"surrogate":                     "{\"v\":\"\xed\xa0\x80\"}",
		"overlong":                      "{\"v\":\"\xc0\x80\"}",
		"literal replacement character": "{\"v\":\"\uFFFD\"}",
		"invalid name":                  "{\"n\xffme\":1}",
	} {
		var want, got any
		if err := json.Unmarshal([]byte(payload), &want); err != nil {
			t.Fatalf("%s: json.Unmarshal() error = %v", name, err)
		}
		if err := jsonwire.Decode([]byte(payload), &got, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true}); err != nil ||
			!reflect.DeepEqual(got, want) {
			t.Fatalf("%s: Decode() = %#v, %v; json.Unmarshal = %#v", name, got, err, want)
		}
	}
	var surrogate map[string]string
	_ = jsonwire.Decode([]byte("{\"v\":\"\xed\xa0\x80\"}"), &surrogate, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true})
	if surrogate["v"] != "\uFFFD\uFFFD\uFFFD" {
		t.Fatalf("surrogate = %q, want three U+FFFD", surrogate["v"])
	}

	type named struct {
		Name string `json:"name"`
		Bad  string `json:"b\uFFFDd"`
	}
	payload := []byte("{\"name\":\"x\xff\",\"b\xffd\":\"y\"}")
	var want, got named
	if err := json.Unmarshal(payload, &want); err != nil {
		t.Fatalf("json.Unmarshal(struct) error = %v", err)
	}
	if err := jsonwire.Decode(payload, &got, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true}); err != nil || got != want {
		t.Fatalf("Decode(struct) = %#v, %v; json.Unmarshal = %#v", got, err, want)
	}
}

func TestReplacedNameCollidesWithLiteralReplacementCharacter(t *testing.T) {
	t.Parallel()

	var target map[string]int
	assertKind(t, jsonwire.Decode([]byte("{\"a\uFFFD\":1,\"a\xff\":2}"), &target, jsonwire.DecodeOptions{
		ReplaceInvalidUTF8: true, DisallowDuplicateNames: true,
	}), wire.ErrValidation)
}

// Raw targets see the bytes before replacement: json.RawMessage and
// json.Unmarshaler implementations receive the invalid input unchanged.
func TestRawTargetsReceiveInvalidUTF8Unchanged(t *testing.T) {
	t.Parallel()

	var target struct {
		Raw json.RawMessage `json:"raw"`
	}
	if err := jsonwire.Decode([]byte("{\"raw\":\"\xff\"}"), &target, jsonwire.DecodeOptions{ReplaceInvalidUTF8: true}); err != nil ||
		string(target.Raw) != "\"\xff\"" {
		t.Fatalf("Decode(RawMessage) = %q, %v", target.Raw, err)
	}
}
