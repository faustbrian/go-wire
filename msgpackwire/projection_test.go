package msgpackwire_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/msgpackwire"
	"github.com/vmihailenco/msgpack/v5"
)

type projectionCallbackValue struct {
	Value int
	Calls int
}

func (value *projectionCallbackValue) UnmarshalMsgpack(payload []byte) error {
	value.Calls++
	return msgpack.Unmarshal(payload, &value.Value)
}

func TestDecodeRejectsProjectedMapKeys(t *testing.T) {
	integerKeys := []byte{0x82, 0xcc, 1, 0xa1, 'a', 0xcd, 0, 1, 0xa1, 'b'}
	stringKeys := []byte{0x82, 0xa1, 'x', 1, 0xc4, 1, 'x', 2}
	t.Run("integer widths", func(t *testing.T) {
		target := map[int]string{9: "retained"}
		before := map[int]string{9: "retained"}
		err := msgpackwire.Decode(integerKeys, &target, msgpackwire.DecodeOptions{})
		if !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, before) {
			t.Fatalf("Decode() = %v, target = %#v; want parse rejection and unchanged target", err, target)
		}
	})
	t.Run("string and binary", func(t *testing.T) {
		target := map[string]int{"retained": 9}
		before := map[string]int{"retained": 9}
		err := msgpackwire.Decode(stringKeys, &target, msgpackwire.DecodeOptions{})
		if !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, before) {
			t.Fatalf("Decode() = %v, target = %#v; want parse rejection and unchanged target", err, target)
		}
	})
	t.Run("nested struct map", func(t *testing.T) {
		type document struct {
			Values map[int]string `msgpack:"values"`
		}
		payload := append([]byte{0x81, 0xa6, 'v', 'a', 'l', 'u', 'e', 's'}, integerKeys...)
		target := document{Values: map[int]string{9: "retained"}}
		before := document{Values: map[int]string{9: "retained"}}
		err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{})
		if !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, before) {
			t.Fatalf("Decode() = %v, target = %#v; want parse rejection and unchanged target", err, target)
		}
	})
	t.Run("explicit last wins", func(t *testing.T) {
		var integers map[int]string
		if err := msgpackwire.Decode(integerKeys, &integers, msgpackwire.DecodeOptions{AllowDuplicateKeys: true}); err != nil || !reflect.DeepEqual(integers, map[int]string{1: "b"}) {
			t.Fatalf("integer opt-in = %#v, %v", integers, err)
		}
		var strings map[string]int
		if err := msgpackwire.Decode(stringKeys, &strings, msgpackwire.DecodeOptions{AllowDuplicateKeys: true}); err != nil || !reflect.DeepEqual(strings, map[string]int{"x": 2}) {
			t.Fatalf("string opt-in = %#v, %v", strings, err)
		}
	})
	t.Run("distinct projected keys", func(t *testing.T) {
		payload := []byte{0x82, 0xcc, 1, 0xa1, 'a', 0xcd, 0, 2, 0xa1, 'b'}
		var target map[int]string
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{}); err != nil || !reflect.DeepEqual(target, map[int]string{1: "a", 2: "b"}) {
			t.Fatalf("distinct keys = %#v, %v", target, err)
		}
	})
}

func TestDecodeProjectionPreservesOptionsAndFieldOwnership(t *testing.T) {
	integerKeys := []byte{0x82, 0xcc, 1, 0xa1, 'a', 0xcd, 0, 1, 0xa1, 'b'}
	t.Run("typed interface preserves widths", func(t *testing.T) {
		var target map[any]string
		if err := msgpackwire.Decode(integerKeys, &target, msgpackwire.DecodeOptions{}); err != nil || !reflect.DeepEqual(target, map[any]string{uint8(1): "a", uint16(1): "b"}) {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
	})
	t.Run("loose interface rejects collapsed widths", func(t *testing.T) {
		target := map[any]string{"retained": "value"}
		err := msgpackwire.Decode(integerKeys, &target, msgpackwire.DecodeOptions{NormalizeNumericWidths: true})
		if !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, map[any]string{"retained": "value"}) {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
		var accepted map[any]string
		if err := msgpackwire.Decode(integerKeys, &accepted, msgpackwire.DecodeOptions{NormalizeNumericWidths: true, AllowDuplicateKeys: true}); err != nil || !reflect.DeepEqual(accepted, map[any]string{uint64(1): "b"}) {
			t.Fatalf("opt-in = %#v, %v", accepted, err)
		}
	})
	t.Run("tagged field ignores unknown projection", func(t *testing.T) {
		type document struct {
			Value int `msgpack:"value"`
		}
		payload := []byte{0x83, 0xa5, 'v', 'a', 'l', 'u', 'e', 7, 0xa1, 'x', 1, 0xc4, 1, 'x', 2}
		var target document
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{}); err != nil || target.Value != 7 {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
	})
	t.Run("noinline nested field", func(t *testing.T) {
		type Nested struct {
			Values map[int]string `msgpack:"values"`
		}
		type document struct {
			Nested `msgpack:"nested,noinline"`
		}
		payload := append([]byte{0x81, 0xa6, 'n', 'e', 's', 't', 'e', 'd', 0x81, 0xa6, 'v', 'a', 'l', 'u', 'e', 's'}, integerKeys...)
		target := document{Nested{Values: map[int]string{9: "retained"}}}
		before := document{Nested{Values: map[int]string{9: "retained"}}}
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, before) {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
	})
	t.Run("custom callback runs once", func(t *testing.T) {
		var target map[string]projectionCallbackValue
		if err := msgpackwire.Decode([]byte{0x81, 0xa1, 'x', 7}, &target, msgpackwire.DecodeOptions{}); err != nil || target["x"] != (projectionCallbackValue{Value: 7, Calls: 1}) {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
	})
}

func TestDecodeProjectionUsesDestinationKeyEquality(t *testing.T) {
	t.Run("array elements", func(t *testing.T) {
		payload := []byte{0x82, 0x91, 0xcc, 1, 0xa1, 'a', 0x91, 0xcd, 0, 1, 0xa1, 'b'}
		target := map[[1]int]string{{9}: "retained"}
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, map[[1]int]string{{9}: "retained"}) {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
	})
	t.Run("float and integer", func(t *testing.T) {
		payload := []byte{0x82, 0xcc, 1, 0xa1, 'a', 0xca, 0x3f, 0x80, 0, 0, 0xa1, 'b'}
		target := map[float64]string{9: "retained"}
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, map[float64]string{9: "retained"}) {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
	})
	t.Run("pointer identity", func(t *testing.T) {
		payload := []byte{0x82, 0xcc, 1, 0xa1, 'a', 0xcd, 0, 1, 0xa1, 'b'}
		var target map[*int]string
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{}); err != nil || len(target) != 2 {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
		values := make(map[string]bool)
		for key, value := range target {
			if key == nil || *key != 1 {
				t.Fatalf("key = %v, want pointer to 1", key)
			}
			values[value] = true
		}
		if !values["a"] || !values["b"] {
			t.Fatalf("values = %#v", values)
		}
	})
	t.Run("error interface identity", func(t *testing.T) {
		payload := []byte{0x82, 0xa1, 'x', 1, 0xc4, 1, 'x', 2}
		var target map[error]int
		if err := msgpackwire.Decode(payload, &target, msgpackwire.DecodeOptions{NormalizeNumericWidths: true}); err != nil || len(target) != 2 {
			t.Fatalf("Decode() = %#v, %v", target, err)
		}
		values := make(map[int]bool)
		for key, value := range target {
			if key == nil || key.Error() != "x" {
				t.Fatalf("key = %v, want independent error identity", key)
			}
			values[value] = true
		}
		if !values[1] || !values[2] {
			t.Fatalf("values = %#v", values)
		}
	})
}
