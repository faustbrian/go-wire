package jsonwire_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/jsonwire"
)

func TestDecodeCanBoundTotalValues(t *testing.T) {
	t.Parallel()

	// The object, both names' values and the array's two items: five values.
	payload := []byte(`{"a":1,"b":[true,null]}`)
	var exact map[string]any
	if err := jsonwire.Decode(payload, &exact, jsonwire.DecodeOptions{MaxValues: 5}); err != nil {
		t.Fatalf("Decode(MaxValues=5) error = %v", err)
	}
	var over map[string]any
	assertKind(t, jsonwire.Decode(payload, &over, jsonwire.DecodeOptions{MaxValues: 4}), wire.ErrSizeLimit)
	if over != nil {
		t.Fatalf("refused payload assigned %#v", over)
	}
	var scalar int
	if err := jsonwire.Decode([]byte(`7`), &scalar, jsonwire.DecodeOptions{MaxValues: 1}); err != nil || scalar != 7 {
		t.Fatalf("Decode(scalar, MaxValues=1) = %d, %v", scalar, err)
	}
}

func TestDecodeCanBoundObjectMembers(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"outer":{"a":1,"b":2,"c":3},"x":0}`)
	var exact map[string]any
	if err := jsonwire.Decode(payload, &exact, jsonwire.DecodeOptions{MaxObjectMembers: 3}); err != nil {
		t.Fatalf("Decode(MaxObjectMembers=3) error = %v", err)
	}
	// The limit applies to each object, including nested ones.
	var over map[string]any
	assertKind(t, jsonwire.Decode(payload, &over, jsonwire.DecodeOptions{MaxObjectMembers: 2}), wire.ErrSizeLimit)
}

func TestDecodeCanBoundNumberLength(t *testing.T) {
	t.Parallel()

	var exact []json.Number
	if err := jsonwire.Decode([]byte(`[-1.5e+10,12345]`), &exact, jsonwire.DecodeOptions{MaxNumberBytes: 9}); err != nil {
		t.Fatalf("Decode(MaxNumberBytes=9) error = %v", err)
	}
	var over []json.Number
	assertKind(t, jsonwire.Decode([]byte(`[1,`+strings.Repeat("9", 10)+`]`), &over, jsonwire.DecodeOptions{MaxNumberBytes: 9}), wire.ErrSizeLimit)
	// Strings and names are not numbers, whatever their content.
	var text map[string]string
	if err := jsonwire.Decode([]byte(`{"1234567890":"1234567890"}`), &text, jsonwire.DecodeOptions{MaxNumberBytes: 1}); err != nil {
		t.Fatalf("Decode(string digits) error = %v", err)
	}
}

func TestDecodeBudgetsRejectNegativeOptions(t *testing.T) {
	t.Parallel()

	for _, options := range []jsonwire.DecodeOptions{{MaxValues: -1}, {MaxObjectMembers: -1}, {MaxNumberBytes: -1}} {
		var target any
		assertKind(t, jsonwire.Decode([]byte(`1`), &target, options), wire.ErrValidation)
		assertKind(t, jsonwire.Validate([]byte(`1`), options), wire.ErrValidation)
	}
}

func TestDecodeCanKeepExactNumbersInUntypedTargets(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"amount":1.10,"huge":1e400,"list":[0.1]}`)
	var exact map[string]any
	if err := jsonwire.Decode(payload, &exact, jsonwire.DecodeOptions{UseNumber: true}); err != nil {
		t.Fatalf("Decode(UseNumber) error = %v", err)
	}
	if exact["amount"] != json.Number("1.10") || exact["huge"] != json.Number("1e400") {
		t.Fatalf("Decode(UseNumber) = %#v", exact)
	}
	if list, ok := exact["list"].([]any); !ok || list[0] != json.Number("0.1") {
		t.Fatalf("Decode(UseNumber) list = %#v", exact["list"])
	}

	var lossy map[string]any
	if err := jsonwire.Decode([]byte(`{"amount":1.10}`), &lossy, jsonwire.DecodeOptions{}); err != nil || lossy["amount"] != 1.1 {
		t.Fatalf("Decode(default) = %#v, %v", lossy, err)
	}
	var typed struct{ Amount float64 }
	if err := jsonwire.Decode([]byte(`{"Amount":2.5}`), &typed, jsonwire.DecodeOptions{UseNumber: true}); err != nil || typed.Amount != 2.5 {
		t.Fatalf("Decode(UseNumber, typed) = %#v, %v", typed, err)
	}
}

func TestValidateChecksOneValueWithoutATarget(t *testing.T) {
	t.Parallel()

	if err := jsonwire.Validate([]byte(` {"a":[1,"x",null]} `), jsonwire.DecodeOptions{}); err != nil {
		t.Fatalf("Validate(valid) error = %v", err)
	}
	if err := jsonwire.Validate([]byte(`1e400`), jsonwire.DecodeOptions{}); err != nil {
		t.Fatalf("Validate(out-of-range number) error = %v", err)
	}
	for _, payload := range []string{``, `{`, `[1,]`, `1 2`, `{"a":1}{}`, "\"\xff\""} {
		assertKind(t, jsonwire.Validate([]byte(payload), jsonwire.DecodeOptions{}), wire.ErrParse)
	}
	assertKind(t, jsonwire.Validate([]byte(`{"a":1,"a":2}`), jsonwire.DecodeOptions{DisallowDuplicateNames: true}), wire.ErrValidation)
	if err := jsonwire.Validate([]byte(`{"a":1,"a":2}`), jsonwire.DecodeOptions{}); err != nil {
		t.Fatalf("Validate(duplicates allowed) error = %v", err)
	}
	assertKind(t, jsonwire.Validate([]byte(`[[1]]`), jsonwire.DecodeOptions{MaxNestedLevels: 1}), wire.ErrSizeLimit)
	assertKind(t, jsonwire.Validate([]byte(`[1,2,3]`), jsonwire.DecodeOptions{MaxValues: 3}), wire.ErrSizeLimit)
	assertKind(t, jsonwire.Validate([]byte(`[1,2]`), jsonwire.DecodeOptions{MaxBytes: 4}), wire.ErrSizeLimit)
	if err := jsonwire.Validate([]byte("\"\xff\""), jsonwire.DecodeOptions{ReplaceInvalidUTF8: true}); err != nil {
		t.Fatalf("Validate(ReplaceInvalidUTF8) error = %v", err)
	}
}

func TestValidateReaderMatchesValidate(t *testing.T) {
	t.Parallel()

	if err := jsonwire.ValidateReader(strings.NewReader(`{"a":1}`), jsonwire.DecodeOptions{MaxObjectMembers: 1}); err != nil {
		t.Fatalf("ValidateReader(valid) error = %v", err)
	}
	assertKind(t, jsonwire.ValidateReader(strings.NewReader(`{"a":1,"b":2}`), jsonwire.DecodeOptions{MaxObjectMembers: 1}), wire.ErrSizeLimit)
	assertKind(t, jsonwire.ValidateReader(nil, jsonwire.DecodeOptions{}), wire.ErrValidation)
}

func TestValidateCapsNestingLikeDecode(t *testing.T) {
	t.Parallel()

	deep := []byte(strings.Repeat("[", 10_001) + strings.Repeat("]", 10_001))
	assertKind(t, jsonwire.Validate(deep, jsonwire.DecodeOptions{}), wire.ErrParse)
	var target any
	assertKind(t, jsonwire.Decode(deep, &target, jsonwire.DecodeOptions{UseNumber: true}), wire.ErrParse)
	limit := []byte(strings.Repeat("[", 10_000) + strings.Repeat("]", 10_000))
	if err := jsonwire.Validate(limit, jsonwire.DecodeOptions{}); err != nil {
		t.Fatalf("Validate(10,000 levels) error = %v", err)
	}
	for _, failing := range [][]byte{[]byte(`[1,2]`), {0xff}} {
		var wireErr *wire.Error
		if err := jsonwire.Validate(failing, jsonwire.DecodeOptions{MaxValues: 1}); !errors.As(err, &wireErr) || wireErr.Op != "validate" {
			t.Fatalf("Validate(%q) error op = %#v", failing, err)
		}
	}
}
