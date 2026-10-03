package msgpackwire

import (
	"reflect"
	"testing"
)

func TestOrdinaryProjectionPreservesRecursiveSourceRefusal(t *testing.T) {
	source := []any{true}
	key, supported := projectedKey(source, reflect.TypeFor[[1]string](), false)
	if key != nil || supported || !reflect.DeepEqual(source, []any{true}) {
		t.Fatalf("recursive incompatible source = %#v, %v, source %#v; want unsupported and unchanged source", key, supported, source)
	}
}
