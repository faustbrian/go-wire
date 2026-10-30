package jsonwire_test

import (
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
				t.Errorf("Decode(levels %d, max %d) error = %v", len(testCase.payload), testCase.maximum, err)
			}
			continue
		}
		assertKind(t, err, wire.ErrSizeLimit)
		if target != nil {
			t.Errorf("Decode(over limit) mutated target = %#v", target)
		}
	}
}

func TestDecodeRejectsNegativeNestedLevels(t *testing.T) {
	t.Parallel()

	var target any
	assertKind(t, jsonwire.Decode([]byte(`1`), &target, jsonwire.DecodeOptions{MaxNestedLevels: -1}), wire.ErrValidation)
}
