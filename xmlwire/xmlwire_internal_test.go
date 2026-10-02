package xmlwire

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire"
)

func TestDefaultCharsetReaderUsesDecodeQuota(t *testing.T) {
	decoder := decoderFor(nil, DecodeOptions{MaxBytes: 3})
	if _, err := decoder.CharsetReader("ascii", strings.NewReader("1234")); !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("configured callback quota: %v", err)
	}
	reader, err := decoder.CharsetReader("latin1", strings.NewReader("\xff\xff\xff"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(reader)
	if err != nil || string(got) != "ÿÿÿ" {
		t.Fatalf("raw-byte inclusive quota: %q, %v", got, err)
	}
	called := false
	custom := decoderFor(nil, DecodeOptions{MaxBytes: 1, CharsetReader: func(_ string, input io.Reader) (io.Reader, error) {
		called = true
		return input, nil
	}})
	if _, err := custom.CharsetReader("vendor", strings.NewReader("1234")); err != nil || !called {
		t.Fatal("custom callback changed")
	}
}

func TestExactBoundaryPredicates(t *testing.T) {
	t.Parallel()

	if exceedsLimit(4, 4) || !exceedsLimit(5, 4) {
		t.Fatal("exceedsLimit() did not preserve the exact boundary")
	}
	for _, value := range []byte{0x80, 0x9f} {
		if !isWindows1252Control(value) {
			t.Fatalf("isWindows1252Control(0x%02x) = false", value)
		}
	}
	for _, value := range []byte{0x7f, 0xa0} {
		if isWindows1252Control(value) {
			t.Fatalf("isWindows1252Control(0x%02x) = true", value)
		}
	}
}

func TestClassifyDecodeErrorRecognizesIndependentParseErrors(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		&xml.SyntaxError{Msg: "malformed", Line: 1},
		charsetError{message: "unsupported"},
	} {
		if classified := classifyDecodeError("decode", err); !errors.Is(classified, wire.ErrParse) {
			t.Fatalf("classifyDecodeError(%T) = %v", err, classified)
		}
	}
}
