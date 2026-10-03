package msgpackwire

import (
	"math"
	"reflect"
)

// admissionBudget owns both passes' conservative work and all structural values
// for one decode. Units are policy estimates, not elapsed time or heap bytes.
type admissionBudget struct {
	values int
	work   int64
}

func newAdmissionBudget(options DecodeOptions) *admissionBudget {
	work := options.MaxKeyComparisonWork
	if work == 0 {
		work = DefaultMaxKeyComparisonWork
	}
	return &admissionBudget{values: defaultLimit(options.MaxTotalValues, DefaultMaxTotalValues), work: work}
}

func (budget *admissionBudget) visit() error {
	if budget.values == 0 {
		return errStructuralLimit
	}
	budget.values--
	return nil
}

func (budget *admissionBudget) reserve(count, units int64) error {
	if count != 0 && units > budget.work/count {
		return errStructuralLimit
	}
	budget.work -= count * units
	return nil
}

func (budget *admissionBudget) reserveMap(length int, keyWeight int64) error {
	n := int64(length)
	// Two passes, each with n*(n-1)/2 comparisons and (n-1)*sum(K).
	if err := budget.reserve(n, n-1); err != nil {
		return err
	}
	if err := budget.reserve(n-1, keyWeight); err != nil {
		return err
	}
	return budget.reserve(n-1, keyWeight)
}

// arrayProjectionUnits covers recursively expanded reflected storage, tuple
// preparation and visits, including zero padding. All arithmetic is capped by
// the remaining allowance before a projected array or tuple is allocated.
func arrayProjectionUnits(target reflect.Type, limit int64) (storage, preparation, comparisons int64, supported, fits bool) {
	if hasCustomProjection(target) {
		return 0, 0, 0, false, true
	}
	if target.Kind() != reflect.Array {
		switch target.Kind() {
		case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			return 1, 1, 0, true, limit >= 1
		default:
			return 0, 0, 0, false, true
		}
	}
	if target.Len() == 0 {
		// No child slots are prepared. Check the same supported-component
		// policy without requiring unused child work to fit this allowance.
		_, _, _, supported, _ := arrayProjectionUnits(target.Elem(), math.MaxInt64)
		return 0, 0, 0, supported, true
	}
	childStorage, childPreparation, childComparisons, supported, fits := arrayProjectionUnits(target.Elem(), limit)
	if !supported || !fits {
		return 0, 0, 0, supported, fits
	}
	n := int64(target.Len())
	multiply := func(units int64) (int64, bool) {
		if n != 0 && units > limit/n {
			return 0, false
		}
		return n * units, true
	}
	storage, fits = multiply(childStorage)
	if !fits {
		return 0, 0, 0, true, false
	}
	children, fits := multiply(childPreparation)
	if !fits || storage > limit-n || children > limit-n-storage {
		return 0, 0, 0, true, false
	}
	preparation = storage + n + children
	// Recursive preparation is at least comparison work. The admitted checked
	// preparation therefore proves this product and sum fit within limit.
	return storage, preparation, n + n*childComparisons, true, true
}

func (budget *admissionBudget) reserveArrayKeys(target reflect.Type, count int) error {
	if target.Kind() != reflect.Array || count == 0 {
		return nil
	}
	_, preparation, comparisons, supported, fits := arrayProjectionUnits(target, budget.work)
	if !supported {
		return nil // Opaque/unsupported projections retain their separate boundary.
	}
	if !fits {
		return errStructuralLimit
	}
	n := int64(count)
	if err := budget.reserve(n, preparation); err != nil {
		return err
	}
	// Each key participates in at most n-1 peer comparisons.
	if n > 1 {
		if comparisons > budget.work/(n-1) {
			return errStructuralLimit
		}
		return budget.reserve(n, (n-1)*comparisons)
	}
	return nil
}
