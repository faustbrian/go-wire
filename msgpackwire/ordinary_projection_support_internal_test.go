package msgpackwire

import (
	"reflect"
	"testing"
)

func TestOrdinaryProjectionDeclinesUnsupportedComponents(t *testing.T) {
	for _, target := range []reflect.Type{
		reflect.TypeFor[[2][0]struct{}](),
		reflect.TypeFor[[0]struct{}](),
		reflect.TypeFor[struct{}](),
		reflect.TypeFor[*int](),
	} {
		t.Run(target.String(), func(t *testing.T) {
			key, supported := projectedKey(nil, target, false)
			if supported || key != nil {
				t.Fatalf("unsupported component projection = %#v, %v; want driver fallback", key, supported)
			}
		})
	}
}

func TestOrdinaryProjectionPreservesEmptyComponentDriverFallback(t *testing.T) {
	var target map[[2][0]struct{}]int
	if err := Decode([]byte{0x81, 0xc0, 1}, &target, DecodeOptions{}); err != nil || !reflect.DeepEqual(target, map[[2][0]struct{}]int{{}: 1}) {
		t.Fatalf("singleton nil-key driver result = %#v, %v", target, err)
	}
	var pointers map[*int]int
	if err := Decode([]byte{0x81, 0xc0, 1}, &pointers, DecodeOptions{}); err != nil || !reflect.DeepEqual(pointers, map[*int]int{nil: 1}) {
		t.Fatalf("singleton nil-pointer driver result = %#v, %v", pointers, err)
	}
}

func TestOrdinaryProjectionPreservesSupportedEmptyArrayComponents(t *testing.T) {
	key, supported := projectedKey(nil, reflect.TypeFor[[2][0]int](), false)
	if !supported || !reflect.DeepEqual(key, []any{[]any{}, []any{}}) {
		t.Fatalf("supported empty components = %#v, %v", key, supported)
	}
}
