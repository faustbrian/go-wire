package xmlwire_test

import (
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire/xmlwire"
)

type charsetProbeReader struct {
	reads int
	cause error
}

func (r *charsetProbeReader) Read([]byte) (int, error) {
	r.reads++
	return 0, r.cause
}

func TestCharsetReaderRejectsLabelsBeforeReading(t *testing.T) {
	for _, label := range []string{"private-unsupported-label", strings.Repeat(" ", 65) + "utf8"} {
		input := &charsetProbeReader{cause: io.EOF}
		_, err := xmlwire.CharsetReader(label, input)
		if err == nil {
			t.Error("invalid label was accepted")
		}
		if input.reads != 0 {
			t.Errorf("invalid label consumed input: %d reads", input.reads)
		}
		if err != nil && strings.Contains(err.Error(), strings.TrimSpace(label)) {
			t.Error("error disclosed rejected label")
		}
	}
}

func TestCharsetReaderReadErrorIsPrivateAndPreservesCause(t *testing.T) {
	cause := errors.New("private-reader-diagnostic")
	_, err := xmlwire.CharsetReader("latin1", &charsetProbeReader{cause: cause})
	if !errors.Is(err, cause) {
		t.Error("reader cause identity was not retained")
	}
	if err == nil || strings.Contains(err.Error(), cause.Error()) {
		t.Error("reader diagnostic was exposed")
	}
}

func TestCharsetReaderWithLimitBoundaries(t *testing.T) {
	for _, maximum := range []int64{0, 3, math.MaxInt64} {
		reader, err := xmlwire.CharsetReaderWithLimit("cp1252", strings.NewReader("\x80\x80\x80"), maximum)
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(reader)
		if err != nil || string(got) != "€€€" {
			t.Fatalf("conversion = %q, %v", got, err)
		}
	}
	input := strings.NewReader("123456789")
	_, err := xmlwire.CharsetReaderWithLimit("ascii", input, 3)
	if !errors.Is(err, xmlwire.ErrPayloadTooLarge) || input.Len() != 5 {
		t.Fatalf("over-limit result = %v, remaining = %d", err, input.Len())
	}
	probe := &charsetProbeReader{cause: io.EOF}
	if _, err := xmlwire.CharsetReaderWithLimit("utf8", probe, -1); err == nil || probe.reads != 0 {
		t.Fatal("negative limit consumed input or was accepted")
	}
	if _, err := xmlwire.CharsetReaderWithLimit("utf8", nil, 3); err == nil {
		t.Fatal("nil input accepted")
	}
	if _, err := xmlwire.CharsetReader(strings.Repeat(" ", 60)+"utf8", strings.NewReader("ok")); err != nil {
		t.Fatalf("exact label limit: %v", err)
	}
}
