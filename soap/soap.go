package soap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/xmlwire"
)

const (
	namespace11 = "http://schemas.xmlsoap.org/soap/envelope/"
	namespace12 = "http://www.w3.org/2003/05/soap-envelope"
)

// DefaultMaxBytes is the default maximum SOAP envelope size.
const DefaultMaxBytes int64 = 1 << 20

// Version identifies a SOAP envelope namespace.
type Version string

const (
	Version11 Version = "1.1"
	Version12 Version = "1.2"
)

// ParseOptions controls SOAP envelope parsing.
type ParseOptions struct {
	MaxBytes      int64
	CharsetReader func(string, io.Reader) (io.Reader, error)
}

// EncodeOptions controls XML serialization for typed SOAP header and body
// values.
type EncodeOptions struct {
	Indent string
}

// FaultReason is one localized SOAP 1.2 fault reason.
type FaultReason struct {
	Language string
	Text     string
}

// Fault is a version-neutral representation of SOAP 1.1 and 1.2 faults.
type Fault struct {
	Version  Version
	Code     string
	Subcodes []string
	Reason   string
	Reasons  []FaultReason
	Actor    string
	Node     string
	Role     string
	Detail   []byte
	Raw      []byte
}

// FaultError reports a valid SOAP fault response.
type FaultError struct {
	Fault Fault
}

func (e *FaultError) Error() string {
	if e.Fault.Reason == "" {
		return "soap fault: " + e.Fault.Code
	}
	return "soap fault: " + e.Fault.Code + ": " + e.Fault.Reason
}

// Unwrap exposes the shared SOAP-fault classification.
func (e *FaultError) Unwrap() error {
	return &wire.Error{Kind: wire.ErrorKindFault, Format: wire.FormatSOAP, Op: "parse fault"}
}

// Envelope is a validated SOAP envelope with preserved raw sections.
type Envelope struct {
	Version Version
	Fault   *Fault

	raw     []byte
	header  []byte
	body    []byte
	options ParseOptions
}

// RawXML returns a copy of the complete source envelope.
func (e *Envelope) RawXML() []byte {
	return bytes.Clone(e.raw)
}

// HeaderXML returns a copy of the exact XML inside the Header element.
func (e *Envelope) HeaderXML() []byte {
	return bytes.Clone(e.header)
}

// BodyXML returns a copy of the exact XML inside the Body element.
func (e *Envelope) BodyXML() []byte {
	return bytes.Clone(e.body)
}

// Parse validates and extracts a SOAP envelope. Valid SOAP faults return both
// the envelope and a *FaultError.
func Parse(payload []byte, options ParseOptions) (*Envelope, error) {
	if options.MaxBytes < 0 {
		return nil, validationError("parse", errors.New("max bytes must not be negative"))
	}
	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	if int64(len(payload)) > maxBytes {
		return nil, validationError("parse", errors.New("payload exceeds size limit"))
	}

	decoder := decoderFor(payload, options)
	root, err := nextStart(decoder)
	if err != nil {
		return nil, parseError("parse envelope", err)
	}
	version, ok := versionForNamespace(root.Name.Space)
	if !ok || root.Name.Local != "Envelope" {
		return nil, envelopeError("validate envelope", fmt.Errorf("unsupported root {%s}%s", root.Name.Space, root.Name.Local))
	}

	envelope := &Envelope{Version: version, raw: bytes.Clone(payload), options: options}
	if err := parseEnvelope(decoder, payload, root, envelope); err != nil {
		return nil, err
	}
	if err := requireDocumentEnd(decoder); err != nil {
		return nil, parseError("parse envelope", err)
	}
	if envelope.Fault != nil {
		return envelope, &FaultError{Fault: *envelope.Fault}
	}

	return envelope, nil
}

// ParseReader reads a bounded stream and parses a SOAP envelope.
func ParseReader(reader io.Reader, options ParseOptions) (*Envelope, error) {
	if reader == nil {
		return nil, validationError("read", errors.New("reader must not be nil"))
	}
	if options.MaxBytes < 0 {
		return nil, validationError("read", errors.New("max bytes must not be negative"))
	}
	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	limit := maxBytes + 1
	if maxBytes == math.MaxInt64 {
		limit = maxBytes
	}
	payload, err := io.ReadAll(io.LimitReader(reader, limit))
	if err != nil {
		return nil, parseError("read", err)
	}
	if int64(len(payload)) > maxBytes {
		return nil, validationError("read", errors.New("payload exceeds size limit"))
	}

	return Parse(payload, options)
}

// DecodeBody decodes the single body child while retaining namespace bindings
// declared on the source envelope.
func (e *Envelope) DecodeBody(target any) error {
	if e.Fault != nil {
		return &FaultError{Fault: *e.Fault}
	}
	if err := validateTarget(target); err != nil {
		return validationError("decode body", err)
	}

	decoder := decoderFor(e.raw, e.options)
	if _, err := nextStart(decoder); err != nil {
		return parseError("decode body", err)
	}
	children := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			return parseError("decode body", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local != "Body" || typed.Name.Space != namespaceForVersion(e.Version) {
				if err := decoder.Skip(); err != nil {
					return parseError("decode body", err)
				}
				continue
			}
			for {
				bodyToken, err := decoder.Token()
				if err != nil {
					return parseError("decode body", err)
				}
				switch bodyTyped := bodyToken.(type) {
				case xml.StartElement:
					children++
					if children == 1 {
						if err := decoder.DecodeElement(target, &bodyTyped); err != nil {
							return validationError("decode body", err)
						}
					} else if err := decoder.Skip(); err != nil {
						return parseError("decode body", err)
					}
				case xml.EndElement:
					if bodyTyped.Name == typed.Name {
						if children != 1 {
							return envelopeError("decode body", fmt.Errorf("body has %d child elements, want 1", children))
						}
						return nil
					}
				case xml.CharData:
					if strings.TrimSpace(string(bodyTyped)) != "" {
						return envelopeError("decode body", errors.New("body contains character data outside its child element"))
					}
				}
			}
		}
	}
}

// Marshal creates a deterministic SOAP envelope around validated raw header
// and body fragments.
func Marshal(version Version, header, body []byte) ([]byte, error) {
	namespace := namespaceForVersion(version)
	if namespace == "" {
		return nil, validationError("marshal", fmt.Errorf("unsupported SOAP version %q", version))
	}
	if err := validateFragment(header); err != nil {
		return nil, validationError("marshal header", err)
	}
	if err := validateFragment(body); err != nil {
		return nil, validationError("marshal body", err)
	}

	var output bytes.Buffer
	output.WriteString(`<soap:Envelope xmlns:soap="`)
	output.WriteString(namespace)
	output.WriteString(`">`)
	if len(header) > 0 {
		output.WriteString(`<soap:Header>`)
		output.Write(header)
		output.WriteString(`</soap:Header>`)
	}
	output.WriteString(`<soap:Body>`)
	output.Write(body)
	output.WriteString(`</soap:Body></soap:Envelope>`)
	return output.Bytes(), nil
}

// MarshalWriter creates a SOAP envelope around raw fragments and writes it to
// writer.
func MarshalWriter(writer io.Writer, version Version, header, body []byte) error {
	payload, err := Marshal(version, header, body)
	if err != nil {
		return err
	}

	return writePayload(writer, payload)
}

// Encode serializes typed header and body values as XML and wraps them in a
// SOAP envelope. A nil header is omitted and a nil body emits an empty Body.
func Encode(version Version, header, body any, options EncodeOptions) ([]byte, error) {
	var headerXML []byte
	if header != nil {
		var err error
		headerXML, err = xmlwire.Encode(header, xmlwire.EncodeOptions{Indent: options.Indent})
		if err != nil {
			return nil, validationError("encode header", err)
		}
	}

	var bodyXML []byte
	if body != nil {
		var err error
		bodyXML, err = xmlwire.Encode(body, xmlwire.EncodeOptions{Indent: options.Indent})
		if err != nil {
			return nil, validationError("encode body", err)
		}
	}

	return Marshal(version, headerXML, bodyXML)
}

// EncodeWriter serializes typed SOAP values and writes the complete envelope
// to writer.
func EncodeWriter(writer io.Writer, version Version, header, body any, options EncodeOptions) error {
	payload, err := Encode(version, header, body, options)
	if err != nil {
		return err
	}

	return writePayload(writer, payload)
}

// MarshalFault validates and serializes a SOAP fault envelope.
func MarshalFault(fault Fault) ([]byte, error) {
	if namespaceForVersion(fault.Version) == "" {
		return nil, validationError("marshal fault", fmt.Errorf("unsupported SOAP version %q", fault.Version))
	}
	if fault.Code == "" {
		return nil, validationError("marshal fault", errors.New("fault code is required"))
	}
	if fault.Version == Version11 && fault.Reason == "" {
		return nil, validationError("marshal fault", errors.New("fault reason is required"))
	}
	if fault.Version == Version12 && len(fault.Reasons) == 0 && fault.Reason == "" {
		return nil, validationError("marshal fault", errors.New("at least one fault reason is required"))
	}
	if err := validateFragment(fault.Detail); err != nil {
		return nil, validationError("marshal fault detail", err)
	}

	var body bytes.Buffer
	if fault.Version == Version11 {
		body.WriteString(`<soap:Fault><faultcode>`)
		writeEscaped(&body, fault.Code)
		body.WriteString(`</faultcode><faultstring>`)
		writeEscaped(&body, fault.Reason)
		body.WriteString(`</faultstring>`)
		writeOptionalElement(&body, "faultactor", fault.Actor)
		if len(fault.Detail) > 0 {
			body.WriteString(`<detail>`)
			body.Write(fault.Detail)
			body.WriteString(`</detail>`)
		}
		body.WriteString(`</soap:Fault>`)
	} else {
		body.WriteString(`<soap:Fault><soap:Code><soap:Value>`)
		writeEscaped(&body, fault.Code)
		body.WriteString(`</soap:Value>`)
		writeSubcodes(&body, fault.Subcodes)
		body.WriteString(`</soap:Code><soap:Reason>`)
		reasons := fault.Reasons
		if len(reasons) == 0 {
			reasons = []FaultReason{{Text: fault.Reason}}
		}
		for _, reason := range reasons {
			body.WriteString(`<soap:Text`)
			if reason.Language != "" {
				body.WriteString(` xml:lang="`)
				writeEscaped(&body, reason.Language)
				body.WriteByte('"')
			}
			body.WriteByte('>')
			writeEscaped(&body, reason.Text)
			body.WriteString(`</soap:Text>`)
		}
		body.WriteString(`</soap:Reason>`)
		writeOptionalElement(&body, "soap:Node", fault.Node)
		writeOptionalElement(&body, "soap:Role", fault.Role)
		if len(fault.Detail) > 0 {
			body.WriteString(`<soap:Detail>`)
			body.Write(fault.Detail)
			body.WriteString(`</soap:Detail>`)
		}
		body.WriteString(`</soap:Fault>`)
	}

	return Marshal(fault.Version, nil, body.Bytes())
}

// MarshalFaultWriter serializes a SOAP fault and writes the complete envelope
// to writer.
func MarshalFaultWriter(writer io.Writer, fault Fault) error {
	payload, err := MarshalFault(fault)
	if err != nil {
		return err
	}

	return writePayload(writer, payload)
}

func writePayload(writer io.Writer, payload []byte) error {
	if writer == nil {
		return validationError("write", errors.New("writer must not be nil"))
	}
	if _, err := io.Copy(writer, bytes.NewReader(payload)); err != nil {
		return &wire.Error{Kind: wire.ErrorKindWrite, Format: wire.FormatSOAP, Op: "write", Err: err}
	}

	return nil
}

type rawFault struct {
	XMLName    xml.Name
	FaultCode  string       `xml:"faultcode"`
	FaultText  string       `xml:"faultstring"`
	FaultActor string       `xml:"faultactor"`
	Detail11   rawDetail    `xml:"detail"`
	Code       rawFaultCode `xml:"Code"`
	Reason     rawReason    `xml:"Reason"`
	Node       string       `xml:"Node"`
	Role       string       `xml:"Role"`
	Detail12   rawDetail    `xml:"Detail"`
}

type rawFaultCode struct {
	Value   string        `xml:"Value"`
	Subcode *rawFaultCode `xml:"Subcode"`
}

type rawReason struct {
	Texts []rawReasonText `xml:"Text"`
}

type rawReasonText struct {
	Language string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	Text     string `xml:",chardata"`
}

type rawDetail struct {
	Inner []byte `xml:",innerxml"`
}

func parseEnvelope(decoder *xml.Decoder, payload []byte, root xml.StartElement, envelope *Envelope) error {
	headerSeen := false
	bodySeen := false
	for {
		token, err := decoder.Token()
		if err != nil {
			return parseError("parse envelope", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Space != root.Name.Space {
				return envelopeError("validate envelope", fmt.Errorf("unexpected child {%s}%s", typed.Name.Space, typed.Name.Local))
			}
			switch typed.Name.Local {
			case "Header":
				if headerSeen || bodySeen {
					return envelopeError("validate envelope", errors.New("header must occur at most once before body"))
				}
				headerSeen = true
				inner, err := captureElement(decoder, payload, typed)
				if err != nil {
					return err
				}
				envelope.header = inner
			case "Body":
				if bodySeen {
					return envelopeError("validate envelope", errors.New("body must occur exactly once"))
				}
				bodySeen = true
				body, fault, err := parseBody(decoder, payload, typed, envelope.Version)
				if err != nil {
					return err
				}
				envelope.body = body
				envelope.Fault = fault
			default:
				return envelopeError("validate envelope", fmt.Errorf("unexpected envelope child %s", typed.Name.Local))
			}
		case xml.EndElement:
			if typed.Name == root.Name {
				if !bodySeen {
					return envelopeError("validate envelope", errors.New("body is required"))
				}
				return nil
			}
		case xml.CharData:
			if strings.TrimSpace(string(typed)) != "" {
				return envelopeError("validate envelope", errors.New("envelope contains character data"))
			}
		}
	}
}

func captureElement(decoder *xml.Decoder, payload []byte, start xml.StartElement) ([]byte, error) {
	contentStart := int(decoder.InputOffset())
	var discard struct{}
	if err := decoder.DecodeElement(&discard, &start); err != nil {
		return nil, parseError("parse envelope", err)
	}
	contentEnd := innerEnd(payload, contentStart, int(decoder.InputOffset()))
	return bytes.Clone(payload[contentStart:contentEnd]), nil
}

func parseBody(decoder *xml.Decoder, payload []byte, start xml.StartElement, version Version) ([]byte, *Fault, error) {
	contentStart := int(decoder.InputOffset())
	elements := 0
	var fault *Fault
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, nil, parseError("parse body", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			elements++
			if typed.Name.Local == "Fault" && typed.Name.Space == namespaceForVersion(version) {
				faultStart := bytes.LastIndexByte(payload[:decoder.InputOffset()], '<')
				var raw rawFault
				if err := decoder.DecodeElement(&raw, &typed); err != nil {
					return nil, nil, parseError("parse fault", err)
				}
				parsed, err := makeFault(version, raw, bytes.Clone(payload[faultStart:decoder.InputOffset()]))
				if err != nil {
					return nil, nil, err
				}
				fault = parsed
			} else if err := decoder.Skip(); err != nil {
				return nil, nil, parseError("parse body", err)
			}
		case xml.EndElement:
			if typed.Name == start.Name {
				contentEnd := innerEnd(payload, contentStart, int(decoder.InputOffset()))
				if fault != nil && elements != 1 {
					return nil, nil, envelopeError("validate fault", errors.New("fault must be the only body child"))
				}
				return bytes.Clone(payload[contentStart:contentEnd]), fault, nil
			}
		case xml.CharData:
			if strings.TrimSpace(string(typed)) != "" {
				return nil, nil, envelopeError("validate body", errors.New("body contains character data"))
			}
		}
	}
}

func innerEnd(payload []byte, contentStart, offset int) int {
	if offset == contentStart {
		return contentStart
	}
	return bytes.LastIndexByte(payload[:offset], '<')
}

func makeFault(version Version, raw rawFault, source []byte) (*Fault, error) {
	fault := &Fault{Version: version, Raw: source}
	if version == Version11 {
		fault.Code = strings.TrimSpace(raw.FaultCode)
		fault.Reason = strings.TrimSpace(raw.FaultText)
		fault.Actor = strings.TrimSpace(raw.FaultActor)
		fault.Detail = bytes.Clone(raw.Detail11.Inner)
	} else {
		fault.Code = strings.TrimSpace(raw.Code.Value)
		for subcode := raw.Code.Subcode; subcode != nil; subcode = subcode.Subcode {
			fault.Subcodes = append(fault.Subcodes, strings.TrimSpace(subcode.Value))
		}
		for _, reason := range raw.Reason.Texts {
			fault.Reasons = append(fault.Reasons, FaultReason{Language: reason.Language, Text: strings.TrimSpace(reason.Text)})
		}
		if len(fault.Reasons) > 0 {
			fault.Reason = fault.Reasons[0].Text
		}
		fault.Node = strings.TrimSpace(raw.Node)
		fault.Role = strings.TrimSpace(raw.Role)
		fault.Detail = bytes.Clone(raw.Detail12.Inner)
	}
	if fault.Code == "" || fault.Reason == "" {
		return nil, envelopeError("validate fault", errors.New("fault code and reason are required"))
	}
	return fault, nil
}

func decoderFor(payload []byte, options ParseOptions) *xml.Decoder {
	decoder := xml.NewDecoder(bytes.NewReader(payload))
	decoder.Strict = true
	decoder.CharsetReader = options.CharsetReader
	if decoder.CharsetReader == nil {
		decoder.CharsetReader = xmlwire.CharsetReader
	}
	return decoder
}

func nextStart(decoder *xml.Decoder) (xml.StartElement, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return xml.StartElement{}, err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			return typed, nil
		case xml.CharData:
			if strings.TrimSpace(string(typed)) != "" {
				return xml.StartElement{}, errors.New("character data before root element")
			}
		}
	}
}

func requireDocumentEnd(decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		switch typed := token.(type) {
		case xml.Comment, xml.ProcInst:
			continue
		case xml.CharData:
			if strings.TrimSpace(string(typed)) == "" {
				continue
			}
		}
		return errors.New("content after envelope")
	}
}

func validateFragment(fragment []byte) error {
	if len(fragment) == 0 {
		return nil
	}
	wrapper := append([]byte(`<root xmlns:soap="`+namespace11+`">`), fragment...)
	wrapper = append(wrapper, []byte(`</root>`)...)
	decoder := xml.NewDecoder(bytes.NewReader(wrapper))
	depth := 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		switch typed := token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 1 && strings.TrimSpace(string(typed)) != "" {
				return errors.New("fragment contains top-level character data")
			}
		}
	}
}

func validateTarget(target any) error {
	if target == nil {
		return errors.New("target must be a non-nil pointer")
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}
	return nil
}

func writeSubcodes(output *bytes.Buffer, subcodes []string) {
	if len(subcodes) == 0 {
		return
	}
	output.WriteString(`<soap:Subcode><soap:Value>`)
	writeEscaped(output, subcodes[0])
	output.WriteString(`</soap:Value>`)
	writeSubcodes(output, subcodes[1:])
	output.WriteString(`</soap:Subcode>`)
}

func writeOptionalElement(output *bytes.Buffer, name, value string) {
	if value == "" {
		return
	}
	output.WriteByte('<')
	output.WriteString(name)
	output.WriteByte('>')
	writeEscaped(output, value)
	output.WriteString(`</`)
	output.WriteString(name)
	output.WriteByte('>')
}

func writeEscaped(output *bytes.Buffer, value string) {
	_ = xml.EscapeText(output, []byte(value))
}

func versionForNamespace(namespace string) (Version, bool) {
	switch namespace {
	case namespace11:
		return Version11, true
	case namespace12:
		return Version12, true
	default:
		return "", false
	}
}

func namespaceForVersion(version Version) string {
	switch version {
	case Version11:
		return namespace11
	case Version12:
		return namespace12
	default:
		return ""
	}
}

func parseError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindParse, Format: wire.FormatSOAP, Op: op, Err: err}
}

func validationError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindValidation, Format: wire.FormatSOAP, Op: op, Err: err}
}

func envelopeError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindEnvelope, Format: wire.FormatSOAP, Op: op, Err: err}
}
