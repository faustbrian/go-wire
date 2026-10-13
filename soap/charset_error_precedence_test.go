package soap_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/soap"
)

type charsetReadCounter struct {
	io.Reader
	reads *int
}

func (reader charsetReadCounter) Read(payload []byte) (int, error) {
	*reader.reads++
	return reader.Reader.Read(payload)
}

// encoding/xml gives a callback error precedence even when a reader is also
// returned. Neither independent SOAP pass may consume that rejected reader.
func TestParseCharsetCallbackErrorPrecedesReturnedReader(t *testing.T) {
	cause := errors.New("private charset constructor failure")
	payload := []byte(`<?xml version="1.0" encoding="custom"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><value/></s:Body></s:Envelope>`)
	for _, test := range []struct {
		name   string
		reader bool
		err    error
	}{
		{"reader and error", true, cause},
		{"nil reader and error", false, cause},
		{"reader without error", true, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			reads := 0
			callback := func(_ string, input io.Reader) (io.Reader, error) {
				if !test.reader {
					return nil, test.err
				}
				return charsetReadCounter{Reader: input, reads: &reads}, test.err
			}
			envelope, err := soap.Parse(payload, soap.ParseOptions{CharsetReader: callback})
			if test.err != nil {
				if envelope != nil || !errors.Is(err, wire.ErrParse) || !errors.Is(err, cause) {
					t.Fatalf("Parse = %v, %v; want no envelope and inspectable parse cause", envelope, err)
				}
				if reads != 0 {
					t.Fatalf("rejected reader was consumed: %d reads", reads)
				}
				// wire.Error already promises categorical text, while its cause
				// stays inspectable and must not be disclosed by applications.
				if strings.Contains(err.Error(), "private") {
					t.Fatal("default error rendered its cause")
				}
				return
			}
			if err != nil || envelope == nil || string(envelope.BodyXML()) != "<value/>" || reads == 0 {
				t.Fatalf("accepted reader = %v, %v; reads %d", envelope, err, reads)
			}
		})
	}
}
