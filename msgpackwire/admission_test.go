package msgpackwire_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/msgpackwire"
)

func TestDecodeAggregateAdmission(t *testing.T) {
	payload := []byte{0x82, 0xa1, 'a', 1, 0xa1, 'b', 2}
	for _, test := range []struct {
		name    string
		options msgpackwire.DecodeOptions
		reject  bool
	}{
		{"nodes below", msgpackwire.DecodeOptions{MaxTotalValues: 4}, true},
		{"nodes exact", msgpackwire.DecodeOptions{MaxTotalValues: 5}, false},
		{"work below", msgpackwire.DecodeOptions{MaxKeyComparisonWork: 25}, true},
		{"work exact", msgpackwire.DecodeOptions{MaxKeyComparisonWork: 26}, false},
		{"defaults", msgpackwire.DecodeOptions{}, false},
		{"duplicate opt-in skips comparisons", msgpackwire.DecodeOptions{AllowDuplicateKeys: true, MaxKeyComparisonWork: 1}, false},
		{"duplicate opt-in retains nodes", msgpackwire.DecodeOptions{AllowDuplicateKeys: true, MaxTotalValues: 4}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := map[string]int{"retained": 9}
			err := msgpackwire.Decode(payload, &target, test.options)
			if test.reject {
				if !errors.Is(err, wire.ErrSizeLimit) || !reflect.DeepEqual(target, map[string]int{"retained": 9}) {
					t.Fatalf("Decode() = %v, %#v; want size rejection and unchanged target", err, target)
				}
			} else if err != nil || target["a"] != 1 || target["b"] != 2 {
				t.Fatalf("Decode() = %v, %#v; want ordinary values", err, target)
			}
		})
	}
	t.Run("sibling maps share node allowance", func(t *testing.T) {
		payload := []byte{0x92, 0x81, 0xa1, 'a', 1, 0x81, 0xa1, 'b', 2}
		for _, limit := range []int{6, 7} {
			target := []map[string]int{{"retained": 9}}
			err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{MaxTotalValues: limit})
			if limit == 6 {
				if !errors.Is(err, wire.ErrSizeLimit) || !reflect.DeepEqual(target, []map[string]int{{"retained": 9}}) {
					t.Fatalf("below boundary = %v, %#v", err, target)
				}
			} else if err != nil || !reflect.DeepEqual(target, []map[string]int{{"retained": 9, "a": 1}, {"b": 2}}) {
				t.Fatalf("exact boundary = %v, %#v", err, target)
			}
		}
	})
	t.Run("sibling maps share comparison allowance", func(t *testing.T) {
		payload := []byte{0x92, 0x82, 0xa1, 'a', 1, 0xa1, 'b', 2, 0x82, 0xa1, 'c', 3, 0xa1, 'd', 4}
		for _, limit := range []int64{51, 52} {
			target := []map[string]int{{"retained": 9}}
			err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{MaxKeyComparisonWork: limit})
			if limit == 51 {
				if !errors.Is(err, wire.ErrSizeLimit) || !reflect.DeepEqual(target, []map[string]int{{"retained": 9}}) {
					t.Fatalf("below shared work = %v, %#v", err, target)
				}
			} else if err != nil || len(target) != 2 || target[0]["a"] != 1 || target[0]["b"] != 2 || target[1]["c"] != 3 || target[1]["d"] != 4 {
				t.Fatalf("exact shared work = %v, %#v", err, target)
			}
		}
	})
	t.Run("refusal precedes destination callbacks", func(t *testing.T) {
		for _, options := range []msgpackwire.DecodeOptions{{MaxTotalValues: 4}, {MaxKeyComparisonWork: 25}} {
			target := admissionCallbackTarget{}
			err := msgpackwire.Decode(payload, &target, options)
			if !errors.Is(err, wire.ErrSizeLimit) || target.Calls != 0 {
				t.Fatalf("Decode() = %v, %#v; want no destination callback or mutation", err, target)
			}
		}
		var callback admissionCallbackTarget
		if err := msgpackwire.Decode(payload, &callback, msgpackwire.DecodeOptions{}); err != nil || callback.Calls != 1 {
			t.Fatalf("accepted destination callback = %v, %#v", err, callback)
		}
		var accepted map[string]projectionCallbackValue
		if err := msgpackwire.Decode(payload, &accepted, msgpackwire.DecodeOptions{}); err != nil || accepted["a"].Calls != 1 || accepted["b"].Calls != 1 {
			t.Fatalf("accepted callbacks = %v, %#v", err, accepted)
		}
	})
}

type admissionCallbackTarget struct {
	Calls int
}

func (target *admissionCallbackTarget) UnmarshalMsgpack(_ []byte) error {
	target.Calls++
	return nil
}

type admissionReader struct {
	*bytes.Reader
	reads int
}

func (reader *admissionReader) Read(payload []byte) (int, error) {
	reader.reads++
	return reader.Reader.Read(payload)
}

func TestDecodeAggregateNegativeOptionsBeforeReading(t *testing.T) {
	for _, options := range []msgpackwire.DecodeOptions{{MaxTotalValues: -1}, {MaxKeyComparisonWork: -1}} {
		reader := &admissionReader{Reader: bytes.NewReader([]byte{1})}
		target := 9
		err := msgpackwire.DecodeReader(reader, &target, options)
		if !errors.Is(err, wire.ErrValidation) || reader.reads != 0 || target != 9 {
			t.Fatalf("DecodeReader() = %v, reads %d, target %d", err, reader.reads, target)
		}
	}
}

func TestDecodeAggregateArrayKeyPreparation(t *testing.T) {
	// Each encoded key has one element, but preparation includes all three
	// destination slots and their padding. Baseline work is 46; extra work is
	// 18 preparation units plus 6 peer-comparison units.
	payload := []byte{0x82, 0x91, 0xcc, 1, 0xa1, 'a', 0x91, 0xcc, 2, 0xa1, 'b'}
	for _, limit := range []int64{69, 70} {
		target := map[[3]int16]string{{9}: "retained"}
		err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{MaxKeyComparisonWork: limit})
		if limit == 69 {
			if !errors.Is(err, wire.ErrSizeLimit) || !reflect.DeepEqual(target, map[[3]int16]string{{9}: "retained"}) {
				t.Fatalf("below expanded boundary = %v, %#v", err, target)
			}
		} else if err != nil || target[[3]int16{1}] != "a" || target[[3]int16{2}] != "b" {
			t.Fatalf("exact expanded boundary = %v, %#v", err, target)
		}
	}
}

func TestDecodeAggregateEmptyArrayKeyAtExhaustedAllowance(t *testing.T) {
	type document struct {
		A map[string]int `msgpack:"A"`
		B map[[0]int]int `msgpack:"B"`
	}
	// Root key comparisons use 26 units, A uses 26, and singleton B's empty
	// array key has no expanded storage, preparation or peer comparisons.
	payload := []byte{0x82, 0xa1, 'A', 0x82, 0xa1, 'a', 1, 0xa1, 'b', 2, 0xa1, 'B', 0x81, 0x90, 3}
	for _, limit := range []int64{51, 52} {
		target := document{A: map[string]int{"retained": 9}, B: map[[0]int]int{{}: 9}}
		err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{MaxKeyComparisonWork: limit})
		if limit == 51 {
			before := document{A: map[string]int{"retained": 9}, B: map[[0]int]int{{}: 9}}
			if !errors.Is(err, wire.ErrSizeLimit) || !reflect.DeepEqual(target, before) {
				t.Fatalf("below inclusive work = %v, %#v", err, target)
			}
		} else if err != nil || target.A["a"] != 1 || target.A["b"] != 2 || target.B[[0]int{}] != 3 {
			t.Fatalf("empty key at exact work = %v, %#v", err, target)
		}
	}
	t.Run("nested empty array components", func(t *testing.T) {
		type nestedDocument struct {
			A map[string]int    `msgpack:"A"`
			B map[[0][3]int]int `msgpack:"B"`
		}
		var target nestedDocument
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{MaxKeyComparisonWork: 52}); err != nil || target.A["a"] != 1 || target.A["b"] != 2 || target.B[[0][3]int{}] != 3 {
			t.Fatalf("nested empty key at exact work = %v, %#v", err, target)
		}
	})
}
