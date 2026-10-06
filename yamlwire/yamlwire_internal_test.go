package yamlwire

import (
	"errors"
	"math"
	"testing"

	"github.com/faustbrian/go-wire/v3/internal/outputlimit"
)

func TestBoundaryPredicates(t *testing.T) {
	t.Parallel()

	for _, indent := range []int{0, 2, 9} {
		if !validIndent(indent) {
			t.Fatalf("validIndent(%d) = false", indent)
		}
	}
	for _, indent := range []int{-1, 1, 10} {
		if validIndent(indent) {
			t.Fatalf("validIndent(%d) = true", indent)
		}
	}
	if hasBlockScalar(-1) || !hasBlockScalar(0) {
		t.Fatal("hasBlockScalar() did not preserve the sentinel boundary")
	}
	if exceedsLimit(4, 4) || !exceedsLimit(5, 4) {
		t.Fatal("exceedsLimit() did not preserve the exact boundary")
	}
	if depthLimitEnabled(0) || !depthLimitEnabled(1) {
		t.Fatal("depthLimitEnabled() did not preserve the zero boundary")
	}
	if exceedsDepth(2, 2) || !exceedsDepth(3, 2) {
		t.Fatal("exceedsDepth() did not preserve the exact boundary")
	}
}

func TestAliasLimitPredicates(t *testing.T) {
	t.Parallel()

	if needsLimitPlugin(DecodeOptions{}) {
		t.Fatal("needsLimitPlugin() enabled an unconfigured plugin")
	}
	for _, options := range []DecodeOptions{{DisallowAliases: true}, {MaxAliases: 1}, {MaxDepth: 1}} {
		if !needsLimitPlugin(options) {
			t.Fatalf("needsLimitPlugin(%+v) = false", options)
		}
	}
	if aliasLimitEnabled(false, 0) || !aliasLimitEnabled(true, 0) || !aliasLimitEnabled(false, 1) {
		t.Fatal("aliasLimitEnabled() did not preserve independent controls")
	}
	if aliasesDisabled(false, 1) || aliasesDisabled(true, 0) || !aliasesDisabled(true, 1) {
		t.Fatal("aliasesDisabled() did not preserve the first-alias boundary")
	}
	if aliasesExceeded(0, 1) || aliasesExceeded(1, 1) || !aliasesExceeded(1, 2) {
		t.Fatal("aliasesExceeded() did not preserve the configured boundary")
	}
}

func TestBlockScalarIndicatorBoundaries(t *testing.T) {
	t.Parallel()

	for input, expected := range map[string]int{
		"x: |\n":    3,
		"|":         0,
		" |":        1,
		"- |\n":     2,
		"-   |\n":   4,
		"x: |+\n":   3,
		"x: >-\r\n": 3,
	} {
		if got := blockScalarHeader([]byte(input)).index; got != expected {
			t.Fatalf("blockScalarIndicator(%q) = %d, want %d", input, got, expected)
		}
	}
	for _, input := range []string{"", "x: x", "xx |", "x:|", "text: 000- >", "- 000- |", "x: |0", "x: |--", "x: | text"} {
		if got := blockScalarHeader([]byte(input)).index; got != -1 {
			t.Fatalf("blockScalarIndicator(%q) = %d, want -1", input, got)
		}
	}
}

func TestAddBlockIndentIndicatorsHonorsExactCapacity(t *testing.T) {
	t.Parallel()

	payload := []byte("text: |-\n  value\n")
	if _, err := addBlockIndentIndicators(payload, 2, int64(len(payload))); !errors.Is(err, outputlimit.ErrLimit) {
		t.Fatalf("addBlockIndentIndicators() over-limit error = %v", err)
	}
	got, err := addBlockIndentIndicators(payload, 2, int64(len(payload)+1))
	if err != nil {
		t.Fatalf("addBlockIndentIndicators() exact limit error = %v", err)
	}
	if string(got) != "text: |2-\n  value\n" {
		t.Fatalf("addBlockIndentIndicators() = %q", got)
	}
}

// Admission outcomes exercise arithmetic and quota policy together without
// materializing giant slices, including native 32-bit integer boundaries.
func TestAdmitOutputCapacityWithoutAllocation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		length, delta int
		quota         int64
		want          int
		reject        bool
	}{
		{0, 0, 1, 0, false},
		{0, 1, 1, 1, false},
		{4, -4, 1, 0, false},
		{4, -5, 1, 0, true},
		{4, 2, 5, 0, true},
		{4, 2, 6, 6, false},
		{4, 2, 7, 6, false},
		{math.MaxInt - 1, 1, math.MaxInt64, math.MaxInt, false},
		{math.MaxInt, 0, math.MaxInt64, math.MaxInt, false},
		{math.MaxInt, 1, math.MaxInt64, 0, true},
		{math.MaxInt - 1, 2, math.MaxInt64, 0, true},
		{math.MaxInt, -1, math.MaxInt64, math.MaxInt - 1, false},
		{4, math.MinInt, math.MaxInt64, 0, true},
		{0, -1, 1, 0, true},
		{int(DefaultMaxBytes) - 1, 0, 0, int(DefaultMaxBytes) - 1, false},
		{int(DefaultMaxBytes), 0, 0, int(DefaultMaxBytes), false},
		{int(DefaultMaxBytes) + 1, 0, 0, 0, true},
		{0, 0, -1, 0, true},
	} {
		got, err := admitOutputCapacity(test.length, test.delta, test.quota)
		if test.reject {
			if got != 0 || !errors.Is(err, outputlimit.ErrLimit) {
				t.Fatalf("admission(%d, %d, %d) = (%d, %v), want rejection", test.length, test.delta, test.quota, got, err)
			}
			continue
		}
		if err != nil || got != test.want {
			t.Fatalf("admission(%d, %d, %d) = (%d, %v), want (%d, nil)", test.length, test.delta, test.quota, got, err, test.want)
		}
	}
}

func TestIntermediateOutputLimitSaturatesWithoutAllocation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ configured, want int64 }{
		{-1, -1},
		{0, 2 * DefaultMaxBytes},
		{1, 2},
		{math.MaxInt64/2 - 1, math.MaxInt64 - 3},
		{math.MaxInt64 / 2, math.MaxInt64 - 1},
		{math.MaxInt64/2 + 1, math.MaxInt64},
		{math.MaxInt64, math.MaxInt64},
	} {
		if got := intermediateOutputLimit(test.configured); got != test.want {
			t.Fatalf("scratch ceiling(%d) = %d, want %d", test.configured, got, test.want)
		}
	}
}
