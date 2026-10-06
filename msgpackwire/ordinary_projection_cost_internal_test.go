package msgpackwire

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestOrdinaryZeroSizedArrayProjectionCostOverflow(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("the metadata width requires a 64-bit int")
	}
	element := reflect.TypeFor[[3][0]int]()
	if element.Size() != 0 {
		t.Fatal("element must occupy zero bytes")
	}
	// Go 1.27.1 ArrayOf creates only a descriptor for this pointer-free,
	// zero-sized element. Its element-loop Equal closure is never invoked.
	// Keep this fixture type-only: never allocate or compare array values.
	width := int64(math.MaxInt64/3) + 1
	target := reflect.ArrayOf(int(width), element)
	if target.Size() != 0 {
		t.Fatal("target must occupy zero bytes")
	}
	storage, preparation, comparisons, supported, fits := arrayProjectionUnits(target, math.MaxInt64)
	if !supported || fits || storage != 0 || preparation != 0 || comparisons != 0 {
		t.Fatalf("overflow cost = %d/%d/%d, supported %v, fits %v", storage, preparation, comparisons, supported, fits)
	}
	budget := &admissionBudget{work: math.MaxInt64}
	if err := budget.reserveArrayKeys(target, 1); !errors.Is(err, errStructuralLimit) || budget.work != math.MaxInt64 {
		t.Fatalf("overflow reservation = %v, remaining %d", err, budget.work)
	}
}

func TestOrdinaryZeroSizedArrayProjectionCostControl(t *testing.T) {
	target := reflect.TypeFor[[2][3][0]int]()
	if target.Size() != 0 {
		t.Fatal("control must occupy zero bytes")
	}
	storage, preparation, comparisons, supported, fits := arrayProjectionUnits(target, math.MaxInt64)
	if !supported || !fits || storage != 0 || preparation != 8 || comparisons != 8 {
		t.Fatalf("control cost = %d/%d/%d, supported %v, fits %v", storage, preparation, comparisons, supported, fits)
	}
	budget := &admissionBudget{work: math.MaxInt64}
	if err := budget.reserveArrayKeys(target, 1); err != nil || budget.work != math.MaxInt64-8 {
		t.Fatalf("control reservation = %v, remaining %d", err, budget.work)
	}
}
