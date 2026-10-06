package msgpackwire_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/msgpackwire"
)

func TestOrdinaryProjectionByteArrayDestinations(t *testing.T) {
	target := map[[2]byte]int{{'z'}: 9}
	collision := []byte{0x82, 0xa1, 'x', 1, 0xc4, 1, 'x', 2}
	if err := msgpackwire.Decode(collision, &target, msgpackwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, map[[2]byte]int{{'z'}: 9}) {
		t.Fatalf("projected byte collision = %v, %#v", err, target)
	}
	distinct := []byte{0x82, 0xa1, 'x', 1, 0xc4, 1, 'y', 2}
	if err := msgpackwire.Decode(distinct, &target, msgpackwire.DecodeOptions{}); err != nil || target[[2]byte{'x'}] != 1 || target[[2]byte{'y'}] != 2 {
		t.Fatalf("distinct byte keys = %v, %#v", err, target)
	}
	if err := msgpackwire.Decode(collision, &target, msgpackwire.DecodeOptions{AllowDuplicateKeys: true}); err != nil || target[[2]byte{'x'}] != 2 {
		t.Fatalf("explicit byte last-wins = %v, %#v", err, target)
	}
}

func TestOrdinaryProjectionNestedDestinationRefusal(t *testing.T) {
	keys := []byte{0x82, 0xcc, 1, 0xa1, 'a', 0xcd, 0, 1, 0xa1, 'b'}
	t.Run("slice map", func(t *testing.T) {
		target := []map[int]string{{9: "retained"}}
		if err := msgpackwire.Decode(append([]byte{0x91}, keys...), &target, msgpackwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, []map[int]string{{9: "retained"}}) {
			t.Fatalf("nested slice refusal = %v, %#v", err, target)
		}
	})
	t.Run("array struct map", func(t *testing.T) {
		type document struct{ Values map[int]string }
		target := document{Values: map[int]string{9: "retained"}}
		if err := msgpackwire.Decode(append([]byte{0x91}, keys...), &target, msgpackwire.DecodeOptions{}); !errors.Is(err, wire.ErrParse) || !reflect.DeepEqual(target, document{Values: map[int]string{9: "retained"}}) {
			t.Fatalf("array-struct refusal = %v, %#v", err, target)
		}
	})
	t.Run("pointer custom callback once", func(t *testing.T) {
		var target **projectionCallbackValue
		if err := msgpackwire.Decode([]byte{7}, &target, msgpackwire.DecodeOptions{}); err != nil || target == nil || *target == nil || (**target).Value != 7 || (**target).Calls != 1 {
			t.Fatalf("pointer custom decode = %v, %#v", err, target)
		}
	})
}

func TestOrdinaryAdmissionRejectsWeightedKeySpan(t *testing.T) {
	// Two one-byte string keys each have two encoded bytes plus four value
	// units. A one-unit allowance cannot admit even the first key's weight.
	target := map[string]int{"retained": 9}
	if err := msgpackwire.Decode([]byte{0x82, 0xa1, 'a', 1, 0xa1, 'b', 2}, &target, msgpackwire.DecodeOptions{MaxKeyComparisonWork: 1}); !errors.Is(err, wire.ErrSizeLimit) || !reflect.DeepEqual(target, map[string]int{"retained": 9}) {
		t.Fatalf("weighted key refusal = %v, %#v", err, target)
	}
}
