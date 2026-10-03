package msgpackwire

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestAdmissionReservationChecksBeforeMultiplication(t *testing.T) {
	budget := &admissionBudget{work: math.MaxInt64}
	if err := budget.reserve(math.MaxInt64, 2); !errors.Is(err, errStructuralLimit) || budget.work != math.MaxInt64 {
		t.Fatalf("overflowing reservation = %v, remaining %d", err, budget.work)
	}
	if err := budget.reserve(1, math.MaxInt64); err != nil || budget.work != 0 {
		t.Fatalf("inclusive reservation = %v, remaining %d", err, budget.work)
	}
}

func TestEmptyArrayProjectionRetainsComponentPolicy(t *testing.T) {
	for _, target := range []reflect.Type{reflect.TypeFor[[0]struct{ Value int }](), reflect.TypeFor[[0]admissionOpaqueElement]()} {
		storage, preparation, comparisons, supported, fits := arrayProjectionUnits(target, 0)
		if supported || !fits || storage != 0 || preparation != 0 || comparisons != 0 {
			t.Fatalf("opaque empty array %v = %d/%d/%d, supported %v, fits %v", target, storage, preparation, comparisons, supported, fits)
		}
	}
}

type admissionOpaqueElement int

func (*admissionOpaqueElement) UnmarshalMsgpack(_ []byte) error { return nil }
