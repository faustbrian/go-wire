package soap

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

// TestRepeatedCaptureReaderDetachesObsoleteSink verifies that a later converter
// can continue reading its predecessor without recording the same bytes twice.
func TestRepeatedCaptureReaderDetachesObsoleteSink(t *testing.T) {
	decoder, source := decoderForCapture(nil, ParseOptions{CharsetReader: func(_ string, input io.Reader) (io.Reader, error) {
		return iotest.OneByteReader(input), nil
	}})
	first, err := decoder.CharsetReader("custom", strings.NewReader("€&amp;<!--keep-->"))
	if err != nil {
		t.Fatal(err)
	}
	obsolete := source.sink
	current, err := decoder.CharsetReader("custom", first)
	if err != nil {
		t.Fatal(err)
	}
	const want = "€&amp;<!--keep-->"
	output := new(strings.Builder)
	count, err := io.Copy(output, current)
	if err != nil || count != int64(len(want)) || output.String() != want {
		t.Fatalf("converted read: count=%d output=%q err=%v", count, output.String(), err)
	}
	if string(source.bytes()) != want {
		t.Fatalf("obsolete reader duplicated capture: %q", source.bytes())
	}
	// io.TeeReader requires a full successful write even when the old capture
	// has been detached; otherwise a live predecessor breaks the current stream.
	payload := []byte("obsolete bytes")
	if count, err := obsolete.Write(payload); count != len(payload) || err != nil {
		t.Fatalf("detached write: count=%d err=%v", count, err)
	}
	if string(source.bytes()) != want || obsolete.buffer != nil {
		t.Fatal("detached sink retained or mutated the active capture")
	}
}
