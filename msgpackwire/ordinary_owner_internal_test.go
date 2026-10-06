package msgpackwire

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestOrdinaryAdmissionOwners(t *testing.T) {
	t.Run("map comparison phases", func(t *testing.T) {
		for _, test := range []struct {
			work, remaining int64
			reject          bool
		}{{1, 1, true}, {7, 5, true}, {14, 0, false}} {
			budget := &admissionBudget{work: test.work}
			err := budget.reserveMap(2, 6) // Two comparisons plus two passes of six units.
			if errors.Is(err, errStructuralLimit) != test.reject || budget.work != test.remaining {
				t.Fatalf("reserveMap(work %d) = %v, remaining %d", test.work, err, budget.work)
			}
		}
	})
	t.Run("expanded array unit boundaries", func(t *testing.T) {
		for _, test := range []struct {
			target                            reflect.Type
			limit                             int64
			storage, preparation, comparisons int64
			supported, fits                   bool
		}{
			{reflect.TypeFor[int](), 0, 1, 1, 0, true, false},
			{reflect.TypeFor[int](), 1, 1, 1, 0, true, true},
			// Empty child arrays require no storage; the two outer tuple slots
			// still contribute preparation and comparison units.
			{reflect.TypeFor[[2][0]int](), 1, 0, 0, 0, true, false},
			{reflect.TypeFor[[2][0]int](), 2, 0, 2, 2, true, true},
			{reflect.TypeFor[[1]int](), 0, 0, 0, 0, true, false},
			{reflect.TypeFor[[2]int](), 1, 0, 0, 0, true, false},
			{reflect.TypeFor[[2]int](), 3, 0, 0, 0, true, false},
			{reflect.TypeFor[[2]int](), 6, 2, 6, 2, true, true},
			{reflect.TypeFor[[2]int](), math.MaxInt64, 2, 6, 2, true, true},
			{reflect.TypeFor[[1]int](), math.MaxInt64, 1, 3, 1, true, true},
			{reflect.TypeFor[[2][2]int](), 10, 0, 0, 0, true, false},
			{reflect.TypeFor[[2][2]int](), 17, 0, 0, 0, true, false},
			{reflect.TypeFor[[2][2]int](), 18, 4, 18, 6, true, true},
			{reflect.TypeFor[[1]struct{ Value int }](), 6, 0, 0, 0, false, true},
		} {
			storage, preparation, comparisons, supported, fits := arrayProjectionUnits(test.target, test.limit)
			if storage != test.storage || preparation != test.preparation || comparisons != test.comparisons || supported != test.supported || fits != test.fits {
				t.Fatalf("units(%v, %d) = %d/%d/%d, %v/%v", test.target, test.limit, storage, preparation, comparisons, supported, fits)
			}
		}
	})
	t.Run("array reservation phases", func(t *testing.T) {
		for _, test := range []struct {
			target          reflect.Type
			count           int
			work, remaining int64
			reject          bool
		}{
			{reflect.TypeFor[[1]int](), 0, 0, 0, false},
			{reflect.TypeFor[int](), 2, 0, 0, false},
			{reflect.TypeFor[[0]int](), 2, 0, 0, false},
			{reflect.TypeFor[[1]struct{ Value int }](), 1, 2, 2, false},
			{reflect.TypeFor[[2]int](), 1, 3, 3, true},
			{reflect.TypeFor[[1]int](), 2, 5, 5, true},
			{reflect.TypeFor[[1]int](), 3, 10, 1, true},
			{reflect.TypeFor[[1]int](), 3, 14, 5, true},
			{reflect.TypeFor[[1]int](), 3, 15, 0, false},
			{reflect.TypeFor[[1]int](), 1, 3, 0, false},
			{reflect.TypeFor[[1]int](), 2, 8, 0, false},
			{reflect.TypeFor[[1]int](), 3, math.MaxInt64, math.MaxInt64 - 15, false},
		} {
			budget := &admissionBudget{work: test.work}
			err := budget.reserveArrayKeys(test.target, test.count)
			if errors.Is(err, errStructuralLimit) != test.reject || budget.work != test.remaining {
				t.Fatalf("reserveArrayKeys(%v,%d,%d) = %v, remaining %d", test.target, test.count, test.work, err, budget.work)
			}
		}
	})
}

func TestOrdinaryProjectionScalarOwners(t *testing.T) {
	// Numeric-fit admission dominates several rejection paths in public Decode.
	// These controls directly assert the owned projection's value/support result.
	for _, test := range []struct {
		name      string
		source    any
		target    reflect.Type
		loose     bool
		want      any
		supported bool
	}{
		{"nil interface", nil, reflect.TypeFor[any](), false, nil, true},
		{"nil scalar", nil, reflect.TypeFor[int](), false, 0, true},
		{"nil scalar array", nil, reflect.TypeFor[[1]int](), false, []any{0}, true},
		{"nil struct", nil, reflect.TypeFor[struct{ Value int }](), false, nil, false},
		{"pointer destination", 1, reflect.TypeFor[*int](), false, nil, false},
		{"nonempty interface", 1, reflect.TypeFor[error](), false, nil, false},
		{"uncomparable interface value", []any{1}, reflect.TypeFor[any](), false, []any{1}, false},
		{"loose signed", int16(2), reflect.TypeFor[any](), true, int64(2), true},
		{"loose float", float32(1.5), reflect.TypeFor[any](), true, float64(1.5), true},
		{"loose binary", []byte("x"), reflect.TypeFor[any](), true, "x", true},
		{"custom component", 2, reflect.TypeFor[admissionOpaqueElement](), false, nil, false},
		{"signed overflow", uint16(128), reflect.TypeFor[int8](), false, nil, false},
		{"unsigned value", uint16(2), reflect.TypeFor[uint8](), false, uint8(2), true},
		{"unsigned negative", int8(-1), reflect.TypeFor[uint8](), false, nil, false},
		{"float unsigned", uint8(2), reflect.TypeFor[float64](), false, float64(2), true},
		{"float32 preserves type", float32(1.5), reflect.TypeFor[float32](), false, float32(1.5), true},
		{"float64 preserves type", float64(1.5), reflect.TypeFor[float64](), false, float64(1.5), true},
		// The pinned driver's float decoder converts Uint64 through int64.
		{"float unsigned signed wrapping", uint64(math.MaxUint64), reflect.TypeFor[float64](), false, float64(-1), true},
		{"float unsupported", "x", reflect.TypeFor[float64](), false, nil, false},
		{"float32 double", float64(1.5), reflect.TypeFor[float32](), false, nil, false},
		{"boolean", true, reflect.TypeFor[bool](), false, true, true},
		{"boolean unsupported", "x", reflect.TypeFor[bool](), false, nil, false},
		{"nil byte array", nil, reflect.TypeFor[[2]byte](), false, [2]byte{}, true},
		{"binary byte array", []byte("x"), reflect.TypeFor[[2]byte](), false, [2]byte{'x'}, true},
		{"string byte array", "x", reflect.TypeFor[[2]byte](), false, [2]byte{'x'}, true},
		{"exact string byte array", "xy", reflect.TypeFor[[2]byte](), false, [2]byte{'x', 'y'}, true},
		{"byte array unsupported", []any{uint8(1)}, reflect.TypeFor[[2]byte](), false, nil, false},
		{"byte array fit", "xy", reflect.TypeFor[[1]byte](), false, nil, false},
		{"array unsupported source", true, reflect.TypeFor[[1]int](), false, nil, false},
		{"array unsupported component", []any{1}, reflect.TypeFor[[1]struct{ Value int }](), false, nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := test.source
			switch source := test.source.(type) {
			case []byte:
				before = append([]byte(nil), source...)
			case []any:
				before = append([]any(nil), source...)
			}
			got, supported := projectedKey(test.source, test.target, test.loose)
			if supported != test.supported || !reflect.DeepEqual(got, test.want) || !reflect.DeepEqual(test.source, before) {
				t.Fatalf("projection = %#v/%v, want %#v/%v", got, supported, test.want, test.supported)
			}
		})
	}
	for _, right := range []any{1, []any{1, 2}} {
		if equalProjectedKeys([]any{1}, right) {
			t.Fatalf("different composite identity compared equal: %#v", right)
		}
	}
}

func TestOrdinaryProjectionKnownFieldRouting(t *testing.T) {
	type destination struct {
		Ignored  int `msgpack:"-"`
		_msgpack struct{}
		private  int
		Values   map[[1]int]int
	}
	target := reflect.TypeFor[destination]()
	fields, supported := projectionFields(target)
	wantFields := []numericStructField{{name: "Values", target: reflect.TypeFor[map[[1]int]int]()}}
	if !supported || !reflect.DeepEqual(fields, wantFields) {
		t.Fatalf("known fields = %#v/%v; want only exported Values", fields, supported)
	}
	source := numericMap{
		{key: "Unknown", value: 9},
		{key: "Values", value: numericMap{
			{key: []any{1}, value: 1},
			{key: []any{2}, value: 2},
		}},
	}
	wantSource := numericMap{
		{key: "Unknown", value: 9},
		{key: "Values", value: numericMap{
			{key: []any{1}, value: 1},
			{key: []any{2}, value: 2},
		}},
	}
	budget := &admissionBudget{work: 8}
	if err := rejectProjectedKeys(source, target, false, budget); err != nil || budget.work != 0 {
		t.Fatalf("known nested map = %v, remaining %d; want accepted at eight units", err, budget.work)
	}
	if !reflect.DeepEqual(source, wantSource) {
		t.Fatal("projection admission changed the borrowed source")
	}
}

func TestOrdinaryProjectionUnsupportedRouting(t *testing.T) {
	type embedded struct{ Value int }
	type inlined struct{ embedded }
	type aliased struct {
		Value int `msgpack:"value,alias:other"`
	}
	type interned struct {
		Value string `msgpack:"value,intern"`
	}
	type shadowed struct {
		First  int `msgpack:"value"`
		Second int `msgpack:"value"`
	}
	for _, target := range []reflect.Type{reflect.TypeFor[inlined](), reflect.TypeFor[aliased](), reflect.TypeFor[interned](), reflect.TypeFor[shadowed]()} {
		fields, supported := projectionFields(target)
		if supported || fields != nil {
			t.Fatalf("ambiguous routing %v = %#v/%v", target, fields, supported)
		}
		budget := &admissionBudget{work: 20}
		if err := rejectProjectedKeys(numericMap{{key: "value", value: 1}}, target, false, budget); err != nil || budget.work != 20 {
			t.Fatalf("unsupported map routing %v = %v, work %d", target, err, budget.work)
		}
		if err := rejectProjectedKeys([]any{1}, target, false, budget); err != nil || budget.work != 20 {
			t.Fatalf("unsupported array routing %v = %v, work %d", target, err, budget.work)
		}
	}
}
