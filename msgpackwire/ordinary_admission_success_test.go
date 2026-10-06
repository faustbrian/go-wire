package msgpackwire_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/faustbrian/go-wire/v3/msgpackwire"
)

func TestOrdinaryMapAdmissionWithLargeWorkAllowance(t *testing.T) {
	for _, test := range []struct {
		name    string
		payload []byte
		work    int64
		want    map[int]int
	}{
		{"singleton", []byte{0x81, 1, 2}, math.MaxInt64, map[int]int{1: 2}},
		{"two pairs", []byte{0x82, 1, 2, 3, 4}, math.MaxInt64, map[int]int{1: 2, 3: 4}},
		{"two pairs exact allowance", []byte{0x82, 1, 2, 3, 4}, 22, map[int]int{1: 2, 3: 4}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got map[int]int
			if err := msgpackwire.Decode(test.payload, &got, msgpackwire.DecodeOptions{MaxKeyComparisonWork: test.work}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("decoded map = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestOrdinaryEmptyArrayResetsStruct(t *testing.T) {
	target := struct{ Value int }{Value: 9}
	if err := msgpackwire.Decode([]byte{0x90}, &target, msgpackwire.DecodeOptions{}); err != nil {
		t.Fatal(err)
	}
	if target != (struct{ Value int }{}) {
		t.Fatalf("decoded empty-array struct = %#v, want zero value", target)
	}
}
