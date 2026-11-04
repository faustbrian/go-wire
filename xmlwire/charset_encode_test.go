package xmlwire_test

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/xmlwire"
)

type charsetRecord struct {
	XMLName xml.Name `xml:"request"`
	City    string   `xml:"City"`
}

func TestEncodeCanWriteLatin1WithAMatchingDeclaration(t *testing.T) {
	t.Parallel()

	encoded, err := xmlwire.Encode(charsetRecord{City: "Hämeenlinna"}, xmlwire.EncodeOptions{Charset: "iso-8859-1", IncludeHeader: true})
	if err != nil {
		t.Fatalf("Encode(ISO-8859-1) error = %v", err)
	}
	want := append([]byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<request><City>H"), 0xe4)
	want = append(want, []byte("meenlinna</City></request>")...)
	if !bytes.Equal(encoded, want) {
		t.Fatalf("Encode(ISO-8859-1) = %q, want %q", encoded, want)
	}

	var decoded charsetRecord
	if err := xmlwire.Decode(encoded, &decoded, xmlwire.DecodeOptions{}); err != nil || decoded.City != "Hämeenlinna" {
		t.Fatalf("Decode(Encode) = %#v, %v", decoded, err)
	}
}

func TestEncodeCanWriteWindows1252(t *testing.T) {
	t.Parallel()

	encoded, err := xmlwire.Encode(charsetRecord{City: "€ – ä"}, xmlwire.EncodeOptions{Charset: "windows-1252"})
	if err != nil {
		t.Fatalf("Encode(Windows-1252) error = %v", err)
	}
	if !bytes.Contains(encoded, []byte{0x80, ' ', 0x96, ' ', 0xe4}) {
		t.Fatalf("Encode(Windows-1252) = %q", encoded)
	}
}

func TestEncodeRejectsCharactersOutsideTheCharset(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		charset, city string
	}{
		{"iso-8859-1", "€"},
		{"us-ascii", "ä"},
		{"windows-1252", "\u0081"},
		{"windows-1252", "Ω"},
	} {
		encoded, err := xmlwire.Encode(charsetRecord{City: test.city}, xmlwire.EncodeOptions{Charset: test.charset})
		assertKind(t, err, wire.ErrValidation)
		if encoded != nil {
			t.Fatalf("Encode(%s, %q) returned output %q", test.charset, test.city, encoded)
		}
	}
}

func TestEncodeCharsetLabelsAndDefaults(t *testing.T) {
	t.Parallel()

	utf8Encoded, err := xmlwire.Encode(charsetRecord{City: "Ω"}, xmlwire.EncodeOptions{Charset: "UTF-8", IncludeHeader: true})
	if err != nil || !bytes.HasPrefix(utf8Encoded, []byte(xml.Header)) || !bytes.Contains(utf8Encoded, []byte("Ω")) {
		t.Fatalf("Encode(UTF-8) = %q, %v", utf8Encoded, err)
	}
	ascii, err := xmlwire.Encode(charsetRecord{City: "Turku"}, xmlwire.EncodeOptions{Charset: "ascii", IncludeHeader: true})
	if err != nil || !bytes.HasPrefix(ascii, []byte("<?xml version=\"1.0\" encoding=\"US-ASCII\"?>\n")) {
		t.Fatalf("Encode(US-ASCII) = %q, %v", ascii, err)
	}
	_, err = xmlwire.Encode(charsetRecord{City: "Turku"}, xmlwire.EncodeOptions{Charset: "koi8-r"})
	assertKind(t, err, wire.ErrValidation)
}

func TestEncodeCharsetMaxBytesCapsReturnedBytes(t *testing.T) {
	t.Parallel()

	value := charsetRecord{City: strings.Repeat("ä", 30)}
	plain, err := xmlwire.Encode(value, xmlwire.EncodeOptions{Charset: "iso-8859-1"})
	if err != nil {
		t.Fatalf("Encode error = %v", err)
	}
	exact := int64(len(plain))
	if encoded, err := xmlwire.Encode(value, xmlwire.EncodeOptions{Charset: "iso-8859-1", MaxBytes: exact}); err != nil || !bytes.Equal(encoded, plain) {
		t.Fatalf("Encode(MaxBytes=%d) = %q, %v", exact, encoded, err)
	}
	_, err = xmlwire.Encode(value, xmlwire.EncodeOptions{Charset: "iso-8859-1", MaxBytes: exact - 1})
	assertKind(t, err, wire.ErrSizeLimit)
	// The euro sign is one Windows-1252 byte but three UTF-8 bytes.
	euros := charsetRecord{City: strings.Repeat("€", 40)}
	euroBytes := int64(len("<request><City></City></request>") + 40)
	encoded, err := xmlwire.Encode(euros, xmlwire.EncodeOptions{Charset: "windows-1252", MaxBytes: euroBytes})
	if err != nil || int64(len(encoded)) != euroBytes {
		t.Fatalf("Encode(Windows-1252) = %d bytes, %v", len(encoded), err)
	}
}

func TestEncodeWindows1252RoundTripsEveryDefinedByte(t *testing.T) {
	t.Parallel()

	for value := 0x20; value <= 0xff; value++ {
		if value == 0x7f || value == 0x81 || value == 0x8d || value == 0x8f || value == 0x90 || value == 0x9d {
			continue
		}
		source := []byte{byte(value)}
		reader, err := xmlwire.CharsetReader("windows-1252", bytes.NewReader(source))
		if err != nil {
			t.Fatalf("CharsetReader(0x%02x) error = %v", value, err)
		}
		decoded, _ := io.ReadAll(reader)
		text := string(decoded)
		if strings.ContainsAny(text, "<>&'\"") {
			continue
		}
		encoded, err := xmlwire.Encode(charsetRecord{City: text}, xmlwire.EncodeOptions{Charset: "windows-1252"})
		if err != nil || !bytes.Contains(encoded, append(append([]byte("<City>"), source...), []byte("</City>")...)) {
			t.Fatalf("Encode(0x%02x) = %q, %v", value, encoded, err)
		}
	}
	for _, character := range []rune{0x81, 0x8d, 0x8f, 0x90, 0x9d, 0x80, 0x9f} {
		_, err := xmlwire.Encode(charsetRecord{City: string(character)}, xmlwire.EncodeOptions{Charset: "windows-1252"})
		assertKind(t, err, wire.ErrValidation)
	}
}

func TestEncodeWriterAndHeaderlessCharsetOutput(t *testing.T) {
	t.Parallel()

	var written bytes.Buffer
	if err := xmlwire.EncodeWriter(&written, charsetRecord{City: "Åbo"}, xmlwire.EncodeOptions{Charset: "latin1"}); err != nil {
		t.Fatalf("EncodeWriter error = %v", err)
	}
	if !bytes.Equal(written.Bytes(), []byte("<request><City>\xc5bo</City></request>")) {
		t.Fatalf("EncodeWriter = %q", written.Bytes())
	}
	var failed bytes.Buffer
	err := xmlwire.EncodeWriter(&failed, charsetRecord{City: "€"}, xmlwire.EncodeOptions{Charset: "latin1"})
	assertKind(t, err, wire.ErrValidation)
	if failed.Len() != 0 {
		t.Fatalf("failed EncodeWriter wrote %q", failed.Bytes())
	}
}
