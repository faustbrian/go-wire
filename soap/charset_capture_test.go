package soap_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/soap"
)

// TestParseCharsetSectionsUseConvertedCoordinates covers both versions, exact
// source capacity, lexical markup preservation and namespace-aware decoding.
func TestParseCharsetSectionsUseConvertedCoordinates(t *testing.T) {
	for _, namespace := range []string{"http://schemas.xmlsoap.org/soap/envelope/", "http://www.w3.org/2003/05/soap-envelope"} {
		for _, spare := range []int{0, 256} {
			t.Run(fmt.Sprintf("%s/cap+%d", namespace, spare), func(t *testing.T) {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Errorf("Parse panicked: %v", recovered)
					}
				}()
				rawValue := strings.Repeat("\x80", 32)
				header := `<h:v a='keep'><!--comment-->` + rawValue + `</h:v>`
				body := `<h:v><![CDATA[` + rawValue + `]]></h:v>`
				raw := `<?xml version="1.0" encoding="windows-1252"?><s:Envelope xmlns:s="` + namespace + `" xmlns:h="urn:value"><s:Header>` + header + `</s:Header><s:Body>` + body + `</s:Body></s:Envelope>`
				payload := make([]byte, len(raw), len(raw)+spare)
				copy(payload, raw)
				for _, reader := range []bool{false, true} {
					var envelope *soap.Envelope
					var err error
					if reader {
						envelope, err = soap.ParseReader(bytes.NewReader(payload), soap.ParseOptions{MaxBytes: int64(len(payload))})
					} else {
						envelope, err = soap.Parse(payload, soap.ParseOptions{MaxBytes: int64(len(payload))})
					}
					if err != nil {
						t.Fatalf("Parse reader=%v: %v", reader, err)
					}
					if !bytes.Equal(envelope.RawXML(), payload) {
						t.Fatal("RawXML changed source bytes")
					}
					if want := strings.ReplaceAll(header, "\x80", "€"); string(envelope.HeaderXML()) != want {
						t.Fatalf("HeaderXML = %q, want %q", envelope.HeaderXML(), want)
					}
					if want := strings.ReplaceAll(body, "\x80", "€"); string(envelope.BodyXML()) != want {
						t.Fatalf("BodyXML = %q, want %q", envelope.BodyXML(), want)
					}
					// A namespace-constrained target proves inherited bindings survive reparse.
					var target struct {
						XMLName xml.Name `xml:"urn:value v"`
						Value   string   `xml:",chardata"`
					}
					if err := envelope.DecodeBody(&target); err != nil || target.Value != strings.Repeat("€", 32) {
						t.Fatalf("DecodeBody = %q, %v", target.Value, err)
					}
				}
			})
		}
	}
}

// TestParseCharsetFaultCapture verifies lexical fault and Detail bytes after
// conversion without losing the original complete envelope or typed reason.
func TestParseCharsetFaultCapture(t *testing.T) {
	for _, version := range []soap.Version{soap.Version11, soap.Version12} {
		namespace := "http://schemas.xmlsoap.org/soap/envelope/"
		fault := `<s:Fault><faultcode>s:Server</faultcode><faultstring>` + strings.Repeat("\x80", 32) + `</faultstring><detail><v>\x80</v></detail></s:Fault>`
		if version == soap.Version12 {
			namespace = "http://www.w3.org/2003/05/soap-envelope"
			fault = `<s:Fault><s:Code><s:Value>s:Receiver</s:Value></s:Code><s:Reason><s:Text xml:lang="en">` + strings.Repeat("\x80", 32) + `</s:Text></s:Reason><s:Detail><v>\x80</v></s:Detail></s:Fault>`
		}
		fault = strings.ReplaceAll(fault, `\x80`, "\x80")
		payload := []byte(`<?xml version="1.0" encoding="windows-1252"?><s:Envelope xmlns:s="` + namespace + `"><s:Body>` + fault + `</s:Body></s:Envelope>`)
		payload = payload[:len(payload):len(payload)]
		envelope, err := soap.Parse(payload, soap.ParseOptions{})
		if !errors.Is(err, wire.ErrSOAPFault) || envelope == nil {
			t.Fatalf("fault parse = %v", err)
		}
		want := strings.ReplaceAll(fault, "\x80", "€")
		if string(envelope.BodyXML()) != want || string(envelope.Fault.Raw) != want || string(envelope.Fault.Detail) != `<v>€</v>` || envelope.Fault.Reason != strings.Repeat("€", 32) {
			t.Fatalf("fault capture = %#v, body %q", envelope.Fault, envelope.BodyXML())
		}
		if !bytes.Equal(envelope.RawXML(), payload) {
			t.Fatal("RawXML changed fault source")
		}
	}
}

// TestParseCustomCharsetCoordinates covers shrinking and expanding callback
// output so no one-to-one original-byte mapping is assumed.
func TestParseCustomCharsetCoordinates(t *testing.T) {
	for _, replacement := range []string{"x", "€€€"} {
		payload := []byte(`<?xml version="1.0" encoding="custom"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><v>MARKER</v></s:Body></s:Envelope>`)
		callback := func(_ string, input io.Reader) (io.Reader, error) {
			raw, err := io.ReadAll(input)
			return strings.NewReader(strings.ReplaceAll(string(raw), "MARKER", replacement)), err
		}
		envelope, err := soap.Parse(payload, soap.ParseOptions{CharsetReader: callback})
		if err != nil {
			t.Fatal(err)
		}
		if want := `<v>` + replacement + `</v>`; string(envelope.BodyXML()) != want {
			t.Fatalf("BodyXML = %q, want %q", envelope.BodyXML(), want)
		}
		var target struct {
			Value string `xml:",chardata"`
		}
		if err := envelope.DecodeBody(&target); err != nil || target.Value != replacement {
			t.Fatalf("DecodeBody = %q, %v", target.Value, err)
		}
	}
}

// TestParseUTF8SectionsKeepLexicalBytes locks the no-conversion path instead of
// permitting token re-encoding to change authored prefixes, quotes or comments.
func TestParseUTF8SectionsKeepLexicalBytes(t *testing.T) {
	header := `<v a='original'><!--comment-->€&amp;</v>`
	body := `<v><![CDATA[€]]></v>`
	payload := []byte(`<?xml version="1.0" encoding="UTF-8"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Header>` + header + `</s:Header><s:Body>` + body + `</s:Body></s:Envelope>`)
	envelope, err := soap.Parse(payload, soap.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if string(envelope.HeaderXML()) != header || string(envelope.BodyXML()) != body || !bytes.Equal(envelope.RawXML(), payload) {
		t.Fatalf("lexical sections changed: %q / %q", envelope.HeaderXML(), envelope.BodyXML())
	}
}

// TestParseCharsetCapturePreservesErrors ensures callback construction and read
// failures remain inspectable but never return a partially captured envelope.
func TestParseCharsetCapturePreservesErrors(t *testing.T) {
	cause := errors.New("private charset failure")
	for _, construct := range []bool{false, true} {
		callback := func(_ string, _ io.Reader) (io.Reader, error) {
			if construct {
				return nil, cause
			}
			return iotest.ErrReader(cause), nil
		}
		payload := []byte(`<?xml version="1.0" encoding="custom"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body/></s:Envelope>`)
		envelope, err := soap.Parse(payload, soap.ParseOptions{CharsetReader: callback})
		if envelope != nil || !errors.Is(err, wire.ErrParse) || !errors.Is(err, cause) {
			t.Fatalf("Parse = %#v, %v", envelope, err)
		}
		if strings.Contains(err.Error(), "private") {
			t.Fatal("charset failure leaked")
		}
	}
}

// TestParseRepeatedCharsetConversionUsesCurrentStream exercises the stdlib's
// repeated encoding-instruction path without assuming the original coordinates.
func TestParseRepeatedCharsetConversionUsesCurrentStream(t *testing.T) {
	header := `<v>` + strings.Repeat("MARKER", 32) + `</v>`
	body := `<v>MARKER</v>`
	payload := []byte(`<?xml version="1.0" encoding="custom"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Header>` + header + `</s:Header><?xml version="1.0" encoding="custom"?><s:Body>` + body + `</s:Body></s:Envelope>`)
	calls := 0
	callback := func(_ string, input io.Reader) (io.Reader, error) {
		calls++
		if calls%2 == 0 {
			return input, nil
		}
		raw, err := io.ReadAll(input)
		return bytes.NewReader(bytes.ReplaceAll(raw, []byte("MARKER"), []byte("€€€€€"))), err
	}
	envelope, err := soap.Parse(payload, soap.ParseOptions{CharsetReader: callback})
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.ReplaceAll(header, "MARKER", "€€€€€"); string(envelope.HeaderXML()) != want {
		t.Fatalf("HeaderXML=%q, want %q", envelope.HeaderXML(), want)
	}
	if string(envelope.BodyXML()) != `<v>€€€€€</v>` {
		t.Fatalf("BodyXML=%q", envelope.BodyXML())
	}
}
