package msgpackwire

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestOrdinaryStructuralMapWithoutBudget(t *testing.T) {
	reader := bytes.NewReader([]byte{0x82, 1, 2, 3, 4})
	decoder := msgpack.NewDecoder(reader)
	limits := structuralLimits{
		maxNestedLevels:  DefaultMaxNestedLevels,
		maxArrayElements: DefaultMaxArrayElements,
		maxMapPairs:      DefaultMaxMapPairs,
		reader:           reader,
	}
	if err := validateMessagePackValue(decoder, limits, 0); err != nil {
		t.Fatal(err)
	}
	if reader.Len() != 0 {
		t.Fatalf("unconsumed input = %d bytes", reader.Len())
	}
	if _, err := decoder.PeekCode(); !errors.Is(err, io.EOF) {
		t.Fatalf("decoder after complete map = %v, want EOF", err)
	}
}
