package bsonwire_test

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/faustbrian/go-wire/v2"
	"github.com/faustbrian/go-wire/v2/bsonwire"
)

func structureDocument(elements ...[]byte) []byte {
	result := make([]byte, 4)
	for _, element := range elements {
		result = append(result, element...)
	}
	result = append(result, 0)
	binary.LittleEndian.PutUint32(result, uint32(len(result)))
	return result
}

func structureScope(scope []byte) []byte {
	value := []byte{0, 0, 0, 0, 1, 0, 0, 0, 0} // total length, empty code string
	value = append(value, scope...)
	binary.LittleEndian.PutUint32(value, uint32(len(value)))
	return structureDocument(append([]byte{0x0f, 's', 0}, value...))
}

type structureRecordingTarget struct{ calls int }

func (target *structureRecordingTarget) UnmarshalBSON([]byte) error {
	target.calls++
	return nil
}

func TestCodeWithScopeDuplicatePolicy(t *testing.T) {
	duplicate := structureDocument([]byte{0x10, 'a', 0, 1, 0, 0, 0}, []byte{0x10, 'a', 0, 2, 0, 0, 0})
	payload := structureScope(duplicate)
	if err := bsonwire.Decode(payload, new(any), bsonwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) {
		t.Error("duplicate scope decode was not rejected as parse failure")
	}
	value := bsonwire.D{{Key: "s", Value: bsonwire.CodeWithScope{Scope: bsonwire.D{{Key: "a", Value: int32(1)}, {Key: "a", Value: int32(2)}}}}}
	if _, err := bsonwire.Encode(value, bsonwire.EncodeOptions{}); !errors.Is(err, wire.ErrValidation) {
		t.Error("duplicate scope encode was not rejected as validation failure")
	}
	encoded, err := bsonwire.Encode(value, bsonwire.EncodeOptions{AllowDuplicateKeys: true})
	if err != nil {
		t.Fatal("valid duplicate opt-in encode failed")
	}
	for _, raw := range [][]byte{payload, encoded} {
		var target bsonwire.D
		if err := bsonwire.Decode(raw, &target, bsonwire.DecodeOptions{AllowDuplicateKeys: true}); err != nil {
			t.Fatal("valid duplicate opt-in decode failed")
		}
		scope, ok := target[0].Value.(bsonwire.CodeWithScope)
		if !ok {
			t.Fatal("scope type lost")
		}
		entries, ok := scope.Scope.(bsonwire.D)
		if !ok || len(entries) != 2 || entries[0].Key != "a" || entries[1].Key != "a" || entries[0].Value != int32(1) || entries[1].Value != int32(2) {
			t.Fatal("duplicate entries lost")
		}
	}
}

func TestMalformedNestedStructureRejectedBeforeDecoderCallback(t *testing.T) {
	broken := []byte{5, 0, 0, 0, 1} // invalid nested document terminator
	truncated := []byte{9, 0, 0, 0, 0x10, 'x', 0, 1, 0}
	invalidArray := structureDocument([]byte{0x10, '1', 0, 1, 0, 0, 0})
	invalidCode := structureScope(structureDocument())
	invalidCode[15] = 1 // empty code string must end with NUL
	for _, payload := range [][]byte{
		structureScope(broken),
		structureDocument(append([]byte{0x03, 'd', 0}, broken...)),
		structureDocument(append([]byte{0x04, 'a', 0}, broken...)),
		structureScope(truncated),
		structureDocument(append([]byte{0x03, 'd', 0}, truncated...)),
		structureDocument(append([]byte{0x04, 'a', 0}, invalidArray...)),
		invalidCode,
	} {
		for _, allow := range []bool{false, true} {
			target := new(structureRecordingTarget)
			err := bsonwire.Decode(payload, target, bsonwire.DecodeOptions{AllowDuplicateKeys: allow})
			var classified *wire.Error
			if !errors.Is(err, wire.ErrParse) || !errors.As(err, &classified) || classified.Err == nil {
				t.Error("malformed nested structure was not rejected with retained cause")
			}
			if target.calls != 0 {
				t.Error("decoder callback ran before structural rejection")
			}
		}
	}
}

func TestMalformedScopeBoundariesRejectedBeforeDecoderCallback(t *testing.T) {
	short := structureDocument([]byte{0x0f, 's', 0, 4, 0, 0, 0})
	truncated := structureScope(structureDocument())
	binary.LittleEndian.PutUint32(truncated[16:20], 6)
	trailing := structureScope(structureDocument())
	binary.LittleEndian.PutUint32(trailing[16:20], 4)
	for name, payload := range map[string][]byte{
		"short scope value":        short,
		"truncated scope document": truncated,
		"trailing scope byte":      trailing,
	} {
		t.Run(name, func(t *testing.T) {
			for _, allow := range []bool{false, true} {
				target := new(structureRecordingTarget)
				err := bsonwire.Decode(payload, target, bsonwire.DecodeOptions{AllowDuplicateKeys: allow})
				var classified *wire.Error
				if !errors.Is(err, wire.ErrParse) || !errors.As(err, &classified) || classified.Err == nil {
					t.Fatal("malformed scope was not rejected with retained cause")
				}
				if err.Error() != "wire: parse failure" || !errors.Is(errors.Unwrap(err), classified.Err) {
					t.Fatal("private default text or inspectable cause was lost")
				}
				if target.calls != 0 {
					t.Fatal("malformed scope reached decoder callback")
				}
			}
		})
	}
}

func assertStructureCallbackBoundary(t *testing.T, payload []byte, rejected bool) {
	t.Helper()
	for _, allow := range []bool{false, true} {
		target := new(structureRecordingTarget)
		err := bsonwire.Decode(payload, target, bsonwire.DecodeOptions{AllowDuplicateKeys: allow})
		if !rejected {
			if err != nil || target.calls != 1 {
				t.Fatal("valid structure did not reach decoder exactly once")
			}
			continue
		}
		var classified *wire.Error
		if !errors.Is(err, wire.ErrParse) || !errors.As(err, &classified) || classified.Err == nil {
			t.Fatal("invalid structure was not rejected with retained parse cause")
		}
		if err.Error() != "wire: parse failure" || !errors.Is(errors.Unwrap(err), classified.Err) {
			t.Fatal("private default text or inspectable cause was lost")
		}
		if target.calls != 0 {
			t.Fatal("invalid structure reached decoder callback")
		}
	}
}

func TestRawStructureResumesSiblingValidation(t *testing.T) {
	first := append([]byte{0x03, 'a', 0}, structureDocument()...)
	second := append([]byte{0x03, 'b', 0}, structureDocument()...)
	assertStructureCallbackBoundary(t, structureDocument(first, second), false)
	second[len(second)-1] = 1 // malformed sibling after a complete valid child
	assertStructureCallbackBoundary(t, structureDocument(first, second), true)
}

func TestScopeCodeLengthBoundsBeforeIndexing(t *testing.T) {
	payload := structureScope(structureDocument())
	assertStructureCallbackBoundary(t, payload, false)
	// The code terminator index would equal the scope value length. Admission
	// must reject the code length before attempting to read that index.
	binary.LittleEndian.PutUint32(payload[11:15], 7)
	assertStructureCallbackBoundary(t, payload, true)
}

func TestScopeRejectsTrailingBytesAfterValidDocument(t *testing.T) {
	scope := structureDocument()
	assertStructureCallbackBoundary(t, structureScope(scope), false)
	// Unlike a scope shorter than its minimum, this valid document would pass
	// child-frame admission if the enclosing exact-boundary check were lost.
	assertStructureCallbackBoundary(t, structureScope(append(scope, 0)), true)
}

func TestRawStructureNestingBoundary(t *testing.T) {
	for _, kind := range []byte{0x03, 0x04, 0x0f} {
		for _, depth := range []int{bsonwire.DefaultMaxNestedLevels, bsonwire.DefaultMaxNestedLevels + 1} {
			payload := structureDocument()
			for range depth {
				if kind == 0x0f {
					payload = structureScope(payload)
				} else {
					payload = structureDocument(append([]byte{kind, '0', 0}, payload...))
				}
			}
			for _, allow := range []bool{false, true} {
				target := new(structureRecordingTarget)
				err := bsonwire.Decode(payload, target, bsonwire.DecodeOptions{AllowDuplicateKeys: allow})
				if depth == bsonwire.DefaultMaxNestedLevels {
					if err != nil || target.calls != 1 {
						t.Fatal("exact nesting boundary rejected")
					}
				} else if !errors.Is(err, wire.ErrSizeLimit) || target.calls != 0 {
					t.Fatal("over-depth structure reached decoder callback")
				}
			}
			_, err := bsonwire.Encode(bsonwire.Raw(payload), bsonwire.EncodeOptions{AllowDuplicateKeys: true})
			if depth == bsonwire.DefaultMaxNestedLevels && err != nil {
				t.Fatal("exact raw encode nesting rejected")
			}
			if depth > bsonwire.DefaultMaxNestedLevels && !errors.Is(err, wire.ErrSizeLimit) {
				t.Fatal("over-depth raw encode accepted")
			}
		}
	}
}
