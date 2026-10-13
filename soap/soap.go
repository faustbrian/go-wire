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

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/internal/outputlimit"
	"github.com/faustbrian/go-wire/v3/xmlwire"
)

const (
	namespace11 = "http://schemas.xmlsoap.org/soap/envelope/"
	namespace12 = "http://www.w3.org/2003/05/soap-envelope"
)

// DefaultMaxBytes is the default maximum SOAP envelope size.
const DefaultMaxBytes int64 = 1 << 20

// ErrPayloadTooLarge identifies SOAP output that exceeds the configured limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

// Version identifies a SOAP envelope namespace.
type Version string

const (
	// Version11 identifies the SOAP 1.1 envelope namespace.
	Version11 Version = "1.1"
	// Version12 identifies the SOAP 1.2 envelope namespace.
	Version12 Version = "1.2"
)

// ParseOptions controls SOAP envelope parsing.
type ParseOptions struct {
	// MaxBytes caps supplied input bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// MaxDepth caps parsed XML element depth; zero selects xmlwire.DefaultMaxDepth; negative
	// values fail.
	MaxDepth int
	// CharsetReader is a borrowed conversion callback used by independent XML passes. Nil uses
	// xmlwire charset admission; it provides no context/deadline API.
	CharsetReader func(string, io.Reader) (io.Reader, error)
}

// EncodeOptions controls XML serialization for typed SOAP header and body
// values.
type EncodeOptions struct {
	// MaxBytes caps returned encoded bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// Indent selects indentation of typed header/body fragments, not envelope
	// canonicalization.
	Indent string
}

// MarshalOptions controls raw SOAP envelope and fault serialization.
type MarshalOptions struct {
	// MaxBytes caps returned encoded bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
}

// FaultReason is one localized SOAP 1.2 fault reason.
type FaultReason struct {
	// Language retains the SOAP 1.2 reason xml:lang value; empty does not negotiate a
	// language.
	Language string
	// Text contains reason text, trimmed when parsed and escaped when emitted.
	Text string
}

// Fault is a mutable version-neutral SOAP 1.1/1.2 projection. Codes remain lexical,
// reasons can disclose peer data, and Raw/Detail are not standalone namespace
// snapshots. MarshalFault interprets fields for Version rather than replaying Raw.
type Fault struct {
	// Version identifies the parsed/emitted SOAP namespace; zero is not a supported version.
	Version Version
	// Code retains lexical fault-code text, not an expanded QName or resolved namespace.
	Code string
	// Subcodes contains SOAP 1.2 lexical subcodes in nesting order.
	Subcodes []string
	// Reason is the SOAP 1.1 reason or first parsed SOAP 1.2 reason; it is sensitive
	// diagnostic text.
	Reason string
	// Reasons contains SOAP 1.2 localized reasons; MarshalFault prefers it over Reason when
	// nonempty.
	Reasons []FaultReason
	// Actor contains the optional SOAP 1.1 actor; empty omits it during marshaling.
	Actor string
	// Node contains the optional SOAP 1.2 node; empty omits it during marshaling.
	Node string
	// Role contains the optional SOAP 1.2 role; empty omits it during marshaling.
	Role string
	// Detail contains mutable UTF-8 inner fault XML, validated as a fragment when marshaled.
	Detail []byte
	// Raw contains mutable captured UTF-8 fault XML and is not replayed by MarshalFault.
	Raw []byte
}

// FaultError reports a valid SOAP fault response.
type FaultError struct {
	// Fault contains a parsed fault, if present; FaultError copies the struct shallowly, so
	// slices alias the envelope fault.
	Fault Fault
}

// Error returns the fixed fault category without disclosing fault fields.
func (e *FaultError) Error() string {
	return "soap fault"
}

// Unwrap exposes the shared SOAP-fault classification.
func (e *FaultError) Unwrap() error {
	return &wire.Error{Kind: wire.ErrorKindFault, Format: wire.FormatSOAP, Op: "parse fault"}
}

// Envelope retains parsed SOAP structure and owned raw bytes. Construct it with
// Parse rather than the zero value; methods require a non-nil receiver. Exported
// Version/Fault remain mutable, so synchronize shared use and do not change the
// parsed version before DecodeBody. FaultError shares nested fault slices.
type Envelope struct {
	// Version identifies the parsed/emitted SOAP namespace; zero is not a supported version.
	Version Version
	// Fault contains a parsed fault, if present; FaultError copies the struct shallowly, so
	// slices alias the envelope fault.
	Fault *Fault

	raw     []byte
	header  []byte
	body    []byte
	options ParseOptions
}

// RawXML returns a copy of the complete source envelope.
func (e *Envelope) RawXML() []byte {
	return bytes.Clone(e.raw)
}

// HeaderXML returns a copy of lexical Header content after UTF-8 charset
// conversion. It does not add inherited namespaces.
func (e *Envelope) HeaderXML() []byte {
	return bytes.Clone(e.header)
}

// BodyXML returns a copy of lexical Body content after UTF-8 charset
// conversion. It does not add inherited namespaces.
func (e *Envelope) BodyXML() []byte {
	return bytes.Clone(e.body)
}

// Parse borrows payload until return and extracts SOAP structure. Valid faults
// return both the envelope and a *FaultError; other classified errors return nil.
// RawXML preserves the original source. Captured Header, Body and Fault XML
// retain lexical bytes from the decoder's UTF-8 stream after charset conversion,
// without re-encoding markup or adding inherited namespaces. MaxBytes caps raw
// input, not custom callback expansion. Custom callbacks are borrowed and own
// their output/work bounds; no context, schema or header-role processing is owned.
func Parse(payload []byte, options ParseOptions) (*Envelope, error) {
	if options.MaxBytes < 0 {
		return nil, validationError("parse", errors.New("max bytes must not be negative"))
	}
	if options.MaxDepth < 0 {
		return nil, validationError("parse", errors.New("max depth must not be negative"))
	}
	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	if exceedsLimit(len(payload), maxBytes) {
		return nil, sizeError("parse", ErrPayloadTooLarge)
	}
	if err := validateTokenDepth(payload, options); err != nil {
		return nil, err
	}

	decoder, source := decoderForCapture(payload, options)
	root, err := nextStart(decoder)
	if err != nil {
		return nil, parseError("parse envelope", err)
	}
	version, ok := versionForNamespace(root.Name.Space)
	if !ok || root.Name.Local != "Envelope" {
		return nil, envelopeError("validate envelope", fmt.Errorf("unsupported root {%s}%s", root.Name.Space, root.Name.Local))
	}

	envelope := &Envelope{Version: version, raw: bytes.Clone(payload), options: options}
	if err := parseEnvelope(decoder, source, root, envelope); err != nil {
		return nil, err
	}
	if err := requireDocumentEnd(decoder); err != nil {
		return nil, parseError("parse envelope", err)
	}
	if envelope.Fault != nil {
		// Keep the envelope available on a protocol fault; this is not parse failure.
		// Copying Fault preserves shared nested slices, not an immutable snapshot.
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
	if options.MaxDepth < 0 {
		return nil, validationError("read", errors.New("max depth must not be negative"))
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
	if exceedsLimit(len(payload), maxBytes) {
		return nil, sizeError("read", ErrPayloadTooLarge)
	}

	return Parse(payload, options)
}

// DecodeBody reparses owned raw XML so the child retains inherited namespace
// bindings. A fault is returned before target validation. A non-nil pointer
// target is required, and it can be mutated before a later child/count/type
// failure; discard it on error. Nil receivers panic and custom codecs are borrowed.
func (e *Envelope) DecodeBody(target any) error {
	if e.Fault != nil {
		return &FaultError{Fault: *e.Fault}
	}
	if err := validateTarget(target); err != nil {
		return &wire.Error{Kind: wire.ErrorKindTarget, Format: wire.FormatSOAP, Op: "decode body", Err: err}
	}

	decoder := decoderFor(e.raw, e.options)
	if _, err := nextStart(decoder); err != nil {
		return parseError("decode body", err)
	}
	for {
		token, err := decoder.Token()
		if err != nil {
			return parseError("decode body", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "Body" && typed.Name.Space == namespaceForVersion(e.Version) {
				return decodeBodyElement(decoder, target)
			}
			if err := decoder.Skip(); err != nil {
				return parseError("decode body", err)
			}
		}
	}
}

func decodeBodyElement(decoder *xml.Decoder, target any) error {
	children := 0
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
			if children != 1 {
				return envelopeError("decode body", fmt.Errorf("body has %d child elements, want 1", children))
			}
			return nil
		case xml.CharData:
			if strings.TrimSpace(string(bodyTyped)) != "" {
				return envelopeError("decode body", errors.New("body contains character data outside its child element"))
			}
		}
	}
}

// Marshal creates a deterministic SOAP envelope around validated raw header
// and body fragments.
func Marshal(version Version, header, body []byte) ([]byte, error) {
	return MarshalWithOptions(version, header, body, MarshalOptions{})
}

// MarshalWithOptions borrows stable fragments, checks their XML syntax under a
// wrapper namespace, and emits a fresh envelope within the complete output quota.
// Empty header is omitted; empty body is allowed. It does not enforce Parse
// depth, single-body-child, schema or header-role rules, nor canonical XML.
// No payload is returned on error; fragment copies precede final output admission.
func MarshalWithOptions(version Version, header, body []byte, options MarshalOptions) ([]byte, error) {
	if options.MaxBytes < 0 {
		return nil, validationError("marshal options", errors.New("max bytes must not be negative"))
	}
	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	if exceedsLimit(len(header), maxBytes) || exceedsLimit(len(body), maxBytes) {
		return nil, marshalSizeError("marshal", outputlimit.ErrLimit)
	}
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

	output, _ := outputlimit.New(options.MaxBytes, DefaultMaxBytes)
	if err := writeString(output, `<soap:Envelope xmlns:soap="`, namespace, `">`); err != nil {
		return nil, marshalSizeError("marshal", err)
	}
	if len(header) > 0 {
		if err := writeString(output, `<soap:Header>`); err != nil {
			return nil, marshalSizeError("marshal", err)
		}
		if _, err := output.Write(header); err != nil {
			return nil, marshalSizeError("marshal", err)
		}
		if err := writeString(output, `</soap:Header>`); err != nil {
			return nil, marshalSizeError("marshal", err)
		}
	}
	if err := writeString(output, `<soap:Body>`); err != nil {
		return nil, marshalSizeError("marshal", err)
	}
	if _, err := output.Write(body); err != nil {
		return nil, marshalSizeError("marshal", err)
	}
	if err := writeString(output, `</soap:Body></soap:Envelope>`); err != nil {
		return nil, marshalSizeError("marshal", err)
	}
	return bytes.Clone(output.Bytes()), nil
}

// MarshalWriter creates a SOAP envelope around raw fragments and writes it to
// writer.
func MarshalWriter(writer io.Writer, version Version, header, body []byte) error {
	return MarshalWriterWithOptions(writer, version, header, body, MarshalOptions{})
}

// MarshalWriterWithOptions creates a size-bounded SOAP envelope and writes it
// to writer only after serialization succeeds.
func MarshalWriterWithOptions(writer io.Writer, version Version, header, body []byte, options MarshalOptions) error {
	payload, err := MarshalWithOptions(version, header, body, options)
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
		headerXML, err = xmlwire.Encode(header, xmlwire.EncodeOptions{MaxBytes: options.MaxBytes, Indent: options.Indent})
		if err != nil {
			return nil, validationError("encode header", err)
		}
	}

	var bodyXML []byte
	if body != nil {
		var err error
		bodyXML, err = xmlwire.Encode(body, xmlwire.EncodeOptions{MaxBytes: options.MaxBytes, Indent: options.Indent})
		if err != nil {
			return nil, validationError("encode body", err)
		}
	}

	return MarshalWithOptions(version, headerXML, bodyXML, MarshalOptions{MaxBytes: options.MaxBytes})
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
	return MarshalFaultWithOptions(fault, MarshalOptions{})
}

// MarshalFaultWithOptions checks version, required code/reasons and Detail XML,
// then emits an owned size-bounded envelope. It escapes typed text but does not
// resolve lexical code prefixes or validate a SOAP schema. Stable input is
// borrowed; recursive subcode work and fragment prework are not heap/time quotas.
func MarshalFaultWithOptions(fault Fault, options MarshalOptions) ([]byte, error) {
	if options.MaxBytes < 0 {
		return nil, validationError("marshal fault options", errors.New("max bytes must not be negative"))
	}
	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	if exceedsLimit(len(fault.Detail), maxBytes) {
		return nil, marshalSizeError("marshal fault", outputlimit.ErrLimit)
	}
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

	body, _ := outputlimit.New(options.MaxBytes, DefaultMaxBytes)
	if fault.Version == Version11 {
		if err := writeString(body, `<soap:Fault><faultcode>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeEscaped(body, fault.Code); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeString(body, `</faultcode><faultstring>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeEscaped(body, fault.Reason); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeString(body, `</faultstring>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeOptionalElement(body, "faultactor", fault.Actor); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if len(fault.Detail) != 0 {
			if err := writeString(body, `<detail>`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if _, err := body.Write(fault.Detail); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if err := writeString(body, `</detail>`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
		}
		if err := writeString(body, `</soap:Fault>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
	} else {
		if err := writeString(body, `<soap:Fault><soap:Code><soap:Value>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeEscaped(body, fault.Code); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeString(body, `</soap:Value>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeSubcodes(body, fault.Subcodes); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeString(body, `</soap:Code><soap:Reason>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		reasons := fault.Reasons
		if len(reasons) == 0 {
			reasons = []FaultReason{{Text: fault.Reason}}
		}
		for _, reason := range reasons {
			if err := writeString(body, `<soap:Text`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if reason.Language != "" {
				if err := writeString(body, ` xml:lang="`); err != nil {
					return nil, marshalSizeError("marshal fault", err)
				}
				if err := writeEscaped(body, reason.Language); err != nil {
					return nil, marshalSizeError("marshal fault", err)
				}
				if err := writeString(body, `"`); err != nil {
					return nil, marshalSizeError("marshal fault", err)
				}
			}
			if err := writeString(body, `>`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if err := writeEscaped(body, reason.Text); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if err := writeString(body, `</soap:Text>`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
		}
		if err := writeString(body, `</soap:Reason>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeOptionalElement(body, "soap:Node", fault.Node); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if err := writeOptionalElement(body, "soap:Role", fault.Role); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
		if len(fault.Detail) != 0 {
			if err := writeString(body, `<soap:Detail>`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if _, err := body.Write(fault.Detail); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
			if err := writeString(body, `</soap:Detail>`); err != nil {
				return nil, marshalSizeError("marshal fault", err)
			}
		}
		if err := writeString(body, `</soap:Fault>`); err != nil {
			return nil, marshalSizeError("marshal fault", err)
		}
	}

	return MarshalWithOptions(fault.Version, nil, body.Bytes(), options)
}

// MarshalFaultWriter serializes a SOAP fault and writes the complete envelope
// to writer.
func MarshalFaultWriter(writer io.Writer, fault Fault) error {
	return MarshalFaultWriterWithOptions(writer, fault, MarshalOptions{})
}

// MarshalFaultWriterWithOptions serializes a size-bounded SOAP fault and
// writes it only after serialization succeeds.
func MarshalFaultWriterWithOptions(writer io.Writer, fault Fault, options MarshalOptions) error {
	payload, err := MarshalFaultWithOptions(fault, options)
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

func parseEnvelope(decoder *xml.Decoder, source *capturedXML, root xml.StartElement, envelope *Envelope) error {
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
				inner, err := captureElement(decoder, source, typed)
				if err != nil {
					return err
				}
				envelope.header = inner
			case "Body":
				if bodySeen {
					return envelopeError("validate envelope", errors.New("body must occur exactly once"))
				}
				bodySeen = true
				body, fault, err := parseBody(decoder, source, typed, envelope.Version)
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

// Capture from the same lexical stream counted by decoder offsets, not from
// original encoded bytes or reserialized tokens.
func captureElement(decoder *xml.Decoder, source *capturedXML, start xml.StartElement) ([]byte, error) {
	contentStart := int(decoder.InputOffset())
	var discard struct{}
	if err := decoder.DecodeElement(&discard, &start); err != nil {
		return nil, parseError("parse envelope", err)
	}
	payload := source.bytes()
	contentEnd := innerEnd(payload, contentStart, int(decoder.InputOffset()))
	return bytes.Clone(payload[contentStart:contentEnd]), nil
}

// Body and fault slices use captured decoder bytes so charset conversion cannot
// shift their offsets into the original source or beyond its capacity.
func parseBody(decoder *xml.Decoder, source *capturedXML, start xml.StartElement, version Version) ([]byte, *Fault, error) {
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
				faultStart := bytes.LastIndexByte(source.bytes()[:decoder.InputOffset()], '<')
				var raw rawFault
				if err := decoder.DecodeElement(&raw, &typed); err != nil {
					return nil, nil, parseError("parse fault", err)
				}
				parsed, err := makeFault(version, raw, bytes.Clone(source.bytes()[faultStart:decoder.InputOffset()]))
				if err != nil {
					return nil, nil, err
				}
				fault = parsed
			} else if err := decoder.Skip(); err != nil {
				return nil, nil, parseError("parse body", err)
			}
		case xml.EndElement:
			if typed.Name == start.Name {
				payload := source.bytes()
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
		if len(fault.Reasons) != 0 {
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

// capturedXML owns converted lexical bytes only when a charset callback runs.
// Otherwise it borrows original input for the duration of Parse.
type capturedXML struct {
	// original is the supplied document, used without conversion.
	original []byte
	// converted retains the composite stream counted by InputOffset.
	converted *bytes.Buffer
	// sink records only the currently active conversion reader.
	sink *captureSink
}

// captureSink can be detached when a conversion wraps its previous input.
// Detached readers must not duplicate bytes or retain obsolete capture buffers.
type captureSink struct {
	// buffer is nil after a later charset switch takes over capture ownership.
	buffer *bytes.Buffer
}

// Write records active bytes and discards bytes from superseded reader layers.
func (sink *captureSink) Write(payload []byte) (int, error) {
	if sink.buffer != nil {
		return sink.buffer.Write(payload)
	}
	return len(payload), nil
}

// bytes returns the current lexical stream, including bytes buffered ahead by
// the decoder. The caller must not retain this alias across another token read.
func (source *capturedXML) bytes() []byte {
	if source.converted != nil {
		return source.converted.Bytes()
	}
	return source.original
}

// decoderForCapture preserves consumed lexical bytes across charset switches.
// The initial bytes.Reader has no read-ahead; each switch keeps the consumed
// prefix and records the returned reader's UTF-8 suffix, not converter input.
// Custom callback expansion remains caller-owned, just as its parsing work is.
func decoderForCapture(payload []byte, options ParseOptions) (*xml.Decoder, *capturedXML) {
	decoder := decoderFor(payload, options)
	source := &capturedXML{original: payload}
	charsetReader := decoder.CharsetReader
	decoder.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		if source.sink != nil {
			source.sink.buffer = nil
		}
		reader, err := charsetReader(label, input)
		if err != nil || reader == nil {
			return reader, err
		}
		// A subsequent encoding instruction may run another callback; its
		// prefix is already converted and cannot be indexed in original input.
		offset := int(decoder.InputOffset())
		if source.converted == nil {
			source.converted = new(bytes.Buffer)
			_, _ = source.converted.Write(source.original[:offset])
		} else {
			source.converted.Truncate(offset)
		}
		source.sink = &captureSink{buffer: source.converted}
		return io.TeeReader(reader, source.sink), nil
	}
	return decoder, source
}

func decoderFor(payload []byte, options ParseOptions) *xml.Decoder {
	decoder := xml.NewDecoder(bytes.NewReader(payload))
	decoder.Strict = true
	decoder.CharsetReader = options.CharsetReader
	if decoder.CharsetReader == nil {
		decoder.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
			return xmlwire.CharsetReaderWithLimit(label, input, options.MaxBytes)
		}
	}
	return decoder
}

func validateTokenDepth(payload []byte, options ParseOptions) error {
	maxDepth := options.MaxDepth
	if maxDepth == 0 {
		maxDepth = xmlwire.DefaultMaxDepth
	}
	decoder := decoderFor(payload, options)
	depth := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			// The main parser retains ownership of syntax and charset errors.
			return nil //nolint:nilerr // syntax errors are classified later
		}
		switch token.(type) {
		case xml.StartElement:
			depth++
			if depth > maxDepth {
				return &wire.Error{
					Kind:   wire.ErrorKindSizeLimit,
					Format: wire.FormatSOAP,
					Op:     "parse envelope",
					Err:    xmlwire.ErrNestingTooDeep,
				}
			}
		case xml.EndElement:
			depth--
		}
	}
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

func writeSubcodes(output io.Writer, subcodes []string) error {
	if len(subcodes) == 0 {
		return nil
	}
	if err := writeString(output, `<soap:Subcode><soap:Value>`); err != nil {
		return err
	}
	if err := writeEscaped(output, subcodes[0]); err != nil {
		return err
	}
	if err := writeString(output, `</soap:Value>`); err != nil {
		return err
	}
	if err := writeSubcodes(output, subcodes[1:]); err != nil {
		return err
	}
	return writeString(output, `</soap:Subcode>`)
}

func writeOptionalElement(output io.Writer, name, value string) error {
	if value == "" {
		return nil
	}
	if err := writeString(output, "<", name, ">"); err != nil {
		return err
	}
	if err := writeEscaped(output, value); err != nil {
		return err
	}
	return writeString(output, `</`, name, ">")
}

func exceedsLimit(length int, maximum int64) bool {
	return int64(length) > maximum
}

func writeEscaped(output io.Writer, value string) error {
	return xml.EscapeText(output, []byte(value))
}

func writeString(output io.Writer, values ...string) error {
	for _, value := range values {
		if _, err := io.WriteString(output, value); err != nil {
			return err
		}
	}
	return nil
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

func marshalSizeError(op string, err error) error {
	return &wire.Error{
		Kind:   wire.ErrorKindSizeLimit,
		Format: wire.FormatSOAP,
		Op:     op,
		Err:    errors.Join(ErrPayloadTooLarge, err),
	}
}

func sizeError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindSizeLimit, Format: wire.FormatSOAP, Op: op, Err: err}
}
