package wire_test

import (
	"encoding/xml"
	"errors"
	"reflect"
	"strings"
	"testing"

	wire "github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/bsonwire"
	"github.com/faustbrian/go-wire/jsonwire"
	"github.com/faustbrian/go-wire/soap"
	"github.com/faustbrian/go-wire/tomlwire"
	"github.com/faustbrian/go-wire/xmlwire"
)

type diagnosticCause struct{ text string }

func (e *diagnosticCause) Error() string { return e.text }

type diagnosticReader struct{ cause error }

func (r diagnosticReader) Read([]byte) (int, error) { return 0, r.cause }

func TestDefaultDiagnosticTextPreservesInspectableCause(t *testing.T) {
	cause := &diagnosticCause{text: "private-diagnostic"}
	err := &wire.Error{Kind: wire.ErrorKindValidation, Format: wire.Format(cause.text), Op: cause.text, Err: cause}
	if got := err.Error(); got != "wire: validation failure" {
		t.Error("default text is not categorical")
	}
	var found *diagnosticCause
	if !errors.Is(err, wire.ErrValidation) || !errors.Is(err, cause) || !errors.As(err, &found) || found != cause {
		t.Fatal("classification or original typed cause lost")
	}
	for _, diagnostic := range []error{errors.Unwrap(err), err.Err} {
		if reflect.TypeOf(diagnostic) != reflect.TypeFor[*diagnosticCause]() || reflect.ValueOf(diagnostic).Pointer() != reflect.ValueOf(cause).Pointer() {
			t.Fatal("direct original cause identity lost")
		}
	}
	if err.Format != wire.Format(cause.text) || err.Op != cause.text {
		t.Fatal("diagnostic fields changed")
	}
	unknown := &wire.Error{Kind: wire.ErrorKind(cause.text), Format: err.Format, Op: err.Op, Err: cause}
	if unknown.Error() != "wire: failure" {
		t.Error("unknown classification exposed diagnostic fields")
	}
}

func TestParserDiagnosticNamesArePrivate(t *testing.T) {
	const secret = "private_field"
	key := append([]byte(secret), 0)
	length := 5 + 2*(1+len(key)+4)
	duplicate := []byte{byte(length), 0, 0, 0}
	for range 2 {
		duplicate = append(duplicate, 0x10)
		duplicate = append(duplicate, key...)
		duplicate = append(duplicate, 1, 0, 0, 0)
	}
	duplicate = append(duplicate, 0)
	cases := []struct {
		name           string
		format         wire.Format
		classification error
		decode         func() error
	}{
		{"JSON", wire.FormatJSON, wire.ErrValidation, func() error {
			return jsonwire.Decode([]byte(`{"`+secret+`":true}`), &struct{}{}, jsonwire.DecodeOptions{DisallowUnknownFields: true})
		}},
		{"TOML", wire.FormatTOML, wire.ErrValidation, func() error {
			return tomlwire.Decode([]byte(secret+" = true\n"), &struct{}{}, tomlwire.DecodeOptions{DisallowUnknownFields: true})
		}},
		{"XML", wire.FormatXML, wire.ErrValidation, func() error {
			return xmlwire.Decode([]byte("<"+secret+"/>"), &struct{}{}, xmlwire.DecodeOptions{ExpectedRoot: xml.Name{Local: "expected"}})
		}},
		{"SOAP", wire.FormatSOAP, wire.ErrEnvelope, func() error {
			_, err := soap.Parse([]byte(`<Envelope xmlns="`+secret+`"><Body/></Envelope>`), soap.ParseOptions{})
			return err
		}},
		{"BSON", wire.FormatBSON, wire.ErrParse, func() error { return bsonwire.Decode(duplicate, new(any), bsonwire.DecodeOptions{}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.decode()
			var classified *wire.Error
			if !errors.Is(err, tc.classification) || !errors.As(err, &classified) || classified.Format != tc.format {
				t.Fatal("parser classification lost")
			}
			if strings.Contains(err.Error(), secret) {
				t.Error("ordinary error exposed peer-controlled name")
			}
			if classified.Err == nil || !strings.Contains(classified.Err.Error(), secret) {
				t.Error("trusted diagnostic cause was discarded")
			}
		})
	}
}

func TestParserReaderDiagnosticsArePrivateAndInspectable(t *testing.T) {
	cause := &diagnosticCause{text: "private-reader-diagnostic"}
	for _, decode := range []func() error{
		func() error {
			return xmlwire.DecodeReader(diagnosticReader{cause}, &struct{}{}, xmlwire.DecodeOptions{})
		},
		func() error { _, err := soap.ParseReader(diagnosticReader{cause}, soap.ParseOptions{}); return err },
	} {
		err := decode()
		var found *diagnosticCause
		if !errors.Is(err, wire.ErrParse) || !errors.Is(err, cause) || !errors.As(err, &found) || found != cause {
			t.Fatal("reader cause or classification lost")
		}
		if strings.Contains(err.Error(), cause.text) {
			t.Error("ordinary error exposed reader diagnostic")
		}
	}
}

func TestParsedSOAPFaultDiagnosticsArePrivateAndInspectable(t *testing.T) {
	const secret = "private_fault"
	payload := []byte(`<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/"><Body><Fault><faultcode>` + secret + `</faultcode><faultstring>` + secret + `</faultstring></Fault></Body></Envelope>`)
	envelope, err := soap.Parse(payload, soap.ParseOptions{})
	var fault *soap.FaultError
	if !errors.Is(err, wire.ErrSOAPFault) || !errors.As(err, &fault) || envelope == nil || envelope.Fault == nil {
		t.Fatal("valid SOAP fault was not retained")
	}
	if fault.Fault.Code != secret || fault.Fault.Reason != secret || envelope.Fault.Code != secret || envelope.Fault.Reason != secret {
		t.Fatal("fault diagnostic fields lost")
	}
	if err.Error() != "soap fault" || envelope.DecodeBody(&struct{}{}).Error() != "soap fault" {
		t.Error("parsed fault text is not categorical")
	}
}
