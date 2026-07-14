package soap_test

import (
	"encoding/xml"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/soap"
)

func TestParseSOAP11EnvelopePreservesRawSections(t *testing.T) {
	t.Parallel()

	payload := readFixture(t, "soap11-response.xml")
	envelope, err := soap.Parse(payload, soap.ParseOptions{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if envelope.Version != soap.Version11 {
		t.Fatalf("Parse() version = %q", envelope.Version)
	}
	if !strings.Contains(string(envelope.HeaderXML()), "request-id") {
		t.Fatalf("HeaderXML() = %q", envelope.HeaderXML())
	}
	if !strings.Contains(string(envelope.BodyXML()), "GetRateResponse") {
		t.Fatalf("BodyXML() = %q", envelope.BodyXML())
	}
	if string(envelope.RawXML()) != string(payload) {
		t.Fatal("RawXML() did not preserve the envelope")
	}
}

func TestParseReaderSupportsBoundedStreams(t *testing.T) {
	t.Parallel()

	payload := readFixture(t, "soap11-response.xml")
	envelope, err := soap.ParseReader(strings.NewReader(string(payload)), soap.ParseOptions{MaxBytes: math.MaxInt64})
	if err != nil || envelope.Version != soap.Version11 {
		t.Fatalf("ParseReader() = %#v, %v", envelope, err)
	}

	for _, tc := range []struct {
		reader  io.Reader
		options soap.ParseOptions
		kind    error
	}{
		{reader: nil, kind: wire.ErrValidation},
		{reader: failingReader{}, kind: wire.ErrParse},
		{reader: strings.NewReader(string(payload)), options: soap.ParseOptions{MaxBytes: -1}, kind: wire.ErrValidation},
		{reader: strings.NewReader(string(payload)), options: soap.ParseOptions{MaxBytes: 3}, kind: wire.ErrValidation},
	} {
		_, err := soap.ParseReader(tc.reader, tc.options)
		assertKind(t, err, tc.kind)
	}
}

func TestEnvelopeRawAccessReturnsCopies(t *testing.T) {
	t.Parallel()

	envelope, err := soap.Parse(readFixture(t, "soap11-response.xml"), soap.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, getter := range []func() []byte{envelope.RawXML, envelope.HeaderXML, envelope.BodyXML} {
		first := getter()
		original := first[0]
		first[0] = 'x'
		if getter()[0] != original {
			t.Fatal("raw getter exposed mutable envelope storage")
		}
	}
}

func TestDecodeBodyRetainsInheritedNamespaces(t *testing.T) {
	t.Parallel()

	envelope, err := soap.Parse(readFixture(t, "soap11-response.xml"), soap.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		XMLName xml.Name `xml:"urn:rates GetRateResponse"`
		Amount  string   `xml:"urn:rates Amount"`
	}
	if err := envelope.DecodeBody(&response); err != nil {
		t.Fatalf("DecodeBody() error = %v", err)
	}
	if response.Amount != "12.50" || response.XMLName.Space != "urn:rates" {
		t.Fatalf("DecodeBody() = %#v", response)
	}
}

func TestParseSOAP12FaultReturnsTypedErrorAndEnvelope(t *testing.T) {
	t.Parallel()

	envelope, err := soap.Parse(readFixture(t, "soap12-fault.xml"), soap.ParseOptions{})
	if envelope == nil || envelope.Fault == nil {
		t.Fatal("Parse() did not return the fault envelope")
	}
	if !errors.Is(err, wire.ErrSOAPFault) {
		t.Fatalf("Parse() error = %v, want SOAP fault", err)
	}
	var faultError *soap.FaultError
	if !errors.As(err, &faultError) {
		t.Fatalf("Parse() error type = %T", err)
	}
	fault := faultError.Fault
	if fault.Version != soap.Version12 || fault.Code != "env:Sender" || fault.Reason != "Invalid shipment" {
		t.Fatalf("fault = %#v", fault)
	}
	if len(fault.Subcodes) != 1 || fault.Subcodes[0] != "rates:InvalidPostalCode" {
		t.Fatalf("fault subcodes = %#v", fault.Subcodes)
	}
	if len(fault.Reasons) != 2 || fault.Reasons[1].Language != "fi" {
		t.Fatalf("fault reasons = %#v", fault.Reasons)
	}
	if !strings.Contains(string(fault.Detail), "PostalCode") || len(fault.Raw) == 0 {
		t.Fatalf("fault raw fields = %#v", fault)
	}
}

func TestParseSOAP11FaultShape(t *testing.T) {
	t.Parallel()

	envelope, err := soap.Parse(readFixture(t, "soap11-fault.xml"), soap.ParseOptions{})
	if !errors.Is(err, wire.ErrSOAPFault) {
		t.Fatalf("Parse() error = %v", err)
	}
	fault := envelope.Fault
	if fault.Code != "soap:Server" || fault.Reason != "Carrier unavailable" || fault.Actor != "rates" {
		t.Fatalf("fault = %#v", fault)
	}
}

func TestParseRejectsInvalidEnvelopeStructure(t *testing.T) {
	t.Parallel()

	tests := []string{
		`<Envelope xmlns="urn:not-soap"><Body/></Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/><env:Body/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Header/><env:Header/><env:Body/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/><env:Header/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><other/><env:Body/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Other/><env:Body/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">text<env:Body/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body><env:Fault/><other/></env:Body></env:Envelope>`,
	}

	for _, payload := range tests {
		_, err := soap.Parse([]byte(payload), soap.ParseOptions{})
		assertKind(t, err, wire.ErrEnvelope)
	}
}

func TestParseRejectsMalformedSizeAndOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		payload []byte
		options soap.ParseOptions
		kind    error
	}{
		{payload: readFixture(t, "malformed.xml"), kind: wire.ErrParse},
		{payload: nil, kind: wire.ErrParse},
		{payload: []byte(`text<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/></env:Envelope>`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Header><x></env:Header><env:Body/></env:Envelope>`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Header/></broken>`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body><x></env:Body></env:Envelope>`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body><env:Fault><faultcode>x</env:Fault></env:Body></env:Envelope>`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/></env:Envelope><extra/>`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/></env:Envelope>text`), kind: wire.ErrParse},
		{payload: []byte(`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/></env:Envelope><`), kind: wire.ErrParse},
		{payload: []byte(`<x/>`), options: soap.ParseOptions{MaxBytes: 2}, kind: wire.ErrValidation},
		{payload: []byte(`<x/>`), options: soap.ParseOptions{MaxBytes: -1}, kind: wire.ErrValidation},
	}
	for _, tt := range tests {
		_, err := soap.Parse(tt.payload, tt.options)
		assertKind(t, err, tt.kind)
	}
}

func TestParseAllowsCommentsAroundEnvelopeSections(t *testing.T) {
	t.Parallel()

	payload := []byte(`<!--before--><env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><!--envelope--><env:Body><!--body--><result/></env:Body></env:Envelope><!--after-->`)
	if _, err := soap.Parse(payload, soap.ParseOptions{}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestDecodeBodyRejectsInvalidShapes(t *testing.T) {
	t.Parallel()

	for _, payload := range []string{
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body/></env:Envelope>`,
		`<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body><one/><two/></env:Body></env:Envelope>`,
	} {
		envelope, err := soap.Parse([]byte(payload), soap.ParseOptions{})
		if err != nil {
			t.Fatal(err)
		}
		var target any
		assertKind(t, envelope.DecodeBody(&target), wire.ErrEnvelope)
	}

	envelope, err := soap.Parse(readFixture(t, "soap11-response.xml"), soap.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertKind(t, envelope.DecodeBody(nil), wire.ErrValidation)
	assertKind(t, envelope.DecodeBody(struct{}{}), wire.ErrValidation)
	var nilTarget *struct{}
	assertKind(t, envelope.DecodeBody(nilTarget), wire.ErrValidation)
	var wrong struct {
		Amount int `xml:"Amount"`
	}
	assertKind(t, envelope.DecodeBody(&wrong), wire.ErrValidation)

	faultEnvelope, faultErr := soap.Parse(readFixture(t, "soap11-fault.xml"), soap.ParseOptions{})
	if !errors.Is(faultErr, wire.ErrSOAPFault) || !errors.Is(faultEnvelope.DecodeBody(&wrong), wire.ErrSOAPFault) {
		t.Fatal("DecodeBody() did not preserve SOAP fault classification")
	}

	var target any
	assertKind(t, (&soap.Envelope{}).DecodeBody(&target), wire.ErrParse)
}

func TestMarshalEnvelopeRoundTripsRawFragments(t *testing.T) {
	t.Parallel()

	header := []byte(`<trace xmlns="urn:trace">abc</trace>`)
	body := []byte(`<GetRate xmlns="urn:rates"><PostalCode>00100</PostalCode></GetRate>`)
	payload, err := soap.Marshal(soap.Version12, header, body)
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := soap.Parse(payload, soap.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Version != soap.Version12 || string(envelope.HeaderXML()) != string(header) || string(envelope.BodyXML()) != string(body) {
		t.Fatalf("round trip = %#v", envelope)
	}

	payload, err = soap.Marshal(soap.Version11, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `<soap:Body></soap:Body>`) || strings.Contains(string(payload), `<soap:Header>`) {
		t.Fatalf("Marshal() = %q", payload)
	}
}

func TestMarshalRejectsVersionAndMalformedFragments(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		version soap.Version
		header  []byte
		body    []byte
	}{
		{version: "unknown"},
		{version: soap.Version11, header: []byte(`<broken>`)},
		{version: soap.Version11, body: []byte(`<broken>`)},
		{version: soap.Version11, body: []byte(`text`)},
	} {
		_, err := soap.Marshal(tc.version, tc.header, tc.body)
		if err == nil {
			t.Fatal("Marshal() error = nil")
		}
	}
}

func TestMarshalFaultRoundTripsBothVersions(t *testing.T) {
	t.Parallel()

	tests := []soap.Fault{
		{Version: soap.Version11, Code: "soap:Server", Reason: "Unavailable", Actor: "rates", Detail: []byte(`<retry>later</retry>`)},
		{Version: soap.Version11, Code: "soap:Client", Reason: "Invalid"},
		{Version: soap.Version12, Code: "env:Sender", Subcodes: []string{"rates:Invalid"}, Reasons: []soap.FaultReason{{Language: "en", Text: "Invalid"}, {Language: "fi", Text: "Virhe"}}, Node: "node", Role: "role", Detail: []byte(`<field>postal_code</field>`)},
		{Version: soap.Version12, Code: "env:Receiver", Reason: "Unavailable"},
	}
	for _, fault := range tests {
		payload, err := soap.MarshalFault(fault)
		if err != nil {
			t.Fatalf("MarshalFault() error = %v", err)
		}
		envelope, err := soap.Parse(payload, soap.ParseOptions{})
		if !errors.Is(err, wire.ErrSOAPFault) {
			t.Fatalf("Parse() error = %v", err)
		}
		if envelope.Fault.Code != fault.Code || envelope.Fault.Version != fault.Version {
			t.Fatalf("round-trip fault = %#v", envelope.Fault)
		}
	}
}

func TestMarshalFaultValidatesRequiredFieldsAndDetail(t *testing.T) {
	t.Parallel()

	for _, fault := range []soap.Fault{
		{},
		{Version: soap.Version11},
		{Version: soap.Version11, Code: "code"},
		{Version: soap.Version12, Code: "code"},
		{Version: soap.Version11, Code: "code", Reason: "reason", Detail: []byte(`<broken>`)},
	} {
		if _, err := soap.MarshalFault(fault); err == nil {
			t.Fatal("MarshalFault() error = nil")
		}
	}
}

func TestFaultErrorWithoutReason(t *testing.T) {
	t.Parallel()

	err := &soap.FaultError{Fault: soap.Fault{Code: "code"}}
	if got := err.Error(); got != "soap fault: code" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(err, wire.ErrSOAPFault) {
		t.Fatal("errors.Is() = false")
	}

	err = &soap.FaultError{Fault: soap.Fault{Code: "code", Reason: "reason"}}
	if got := err.Error(); got != "soap fault: code: reason" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestParseRejectsInvalidFaultShapes(t *testing.T) {
	t.Parallel()

	validFault := `<env:Fault><faultcode>env:Server</faultcode><faultstring>failure</faultstring></env:Fault>`
	for _, body := range []string{
		`<env:Fault/>`,
		validFault + `<other/>`,
		`text`,
	} {
		payload := `<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"><env:Body>` + body + `</env:Body></env:Envelope>`
		_, err := soap.Parse([]byte(payload), soap.ParseOptions{})
		assertKind(t, err, wire.ErrEnvelope)
	}
}

func FuzzParse(f *testing.F) {
	f.Add(readFixture(f, "soap11-response.xml"))
	f.Add(readFixture(f, "malformed.xml"))
	f.Fuzz(func(t *testing.T, payload []byte) {
		_, _ = soap.Parse(payload, soap.ParseOptions{MaxBytes: 64 << 10})
	})
}

func BenchmarkParse(b *testing.B) {
	payload := readFixture(b, "soap11-response.xml")
	b.ReportAllocs()
	for b.Loop() {
		if _, err := soap.Parse(payload, soap.ParseOptions{}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	body := []byte(`<GetRate xmlns="urn:rates"><PostalCode>00100</PostalCode></GetRate>`)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := soap.Marshal(soap.Version11, nil, body); err != nil {
			b.Fatal(err)
		}
	}
}

func readFixture(tb testing.TB, name string) []byte {
	tb.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		tb.Fatal(err)
	}
	return payload
}

func assertKind(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want errors.Is(_, %v)", err, target)
	}
	var wireErr *wire.Error
	if !errors.As(err, &wireErr) || wireErr.Format != wire.FormatSOAP {
		t.Fatalf("error = %#v, want SOAP *wire.Error", err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("vendor stream failed")
}
