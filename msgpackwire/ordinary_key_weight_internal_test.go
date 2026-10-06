package msgpackwire

import (
	"errors"
	"math"
	"testing"
)

func TestOrdinaryKeyWeightCheckedAdmission(t *testing.T) {
	for _, test := range []struct {
		name              string
		work, span, nodes int64
		want              int64
		refused           bool
	}{
		{"below exact", 5, 2, 1, 0, true},
		{"exact", 6, 2, 1, 6, false},
		{"above exact", 7, 2, 1, 6, false},
		{"multiplication exceeds allowance", math.MaxInt64, math.MaxInt64/4 + 1, math.MaxInt64/4 + 1, 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := accumulateKeyWeight(test.work, 0, test.span, test.nodes)
			if test.refused {
				if !errors.Is(err, errStructuralLimit) || got != test.want {
					t.Fatalf("refused weight = %d, %v; want unchanged accumulator and structural limit", got, err)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("admitted weight = %d, %v; want %d", got, err, test.want)
			}
		})
	}
}
