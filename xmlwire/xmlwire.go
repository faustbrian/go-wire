package xmlwire

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/internal/outputlimit"
	"github.com/faustbrian/go-wire/v3/internal/valuecheck"
)

const (
	// DefaultMaxBytes is the default maximum XML document size.
	DefaultMaxBytes int64 = 1 << 20
	// DefaultMaxDepth is the default maximum nested XML element depth.
	DefaultMaxDepth = 1000
)

// ErrPayloadTooLarge identifies XML documents over the configured limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

// ErrNestingTooDeep identifies XML documents over the configured depth limit.
var ErrNestingTooDeep = errors.New("XML nesting exceeds depth limit")

// DecodeOptions controls XML parsing and validation.
type DecodeOptions struct {
	// MaxBytes caps supplied input bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// MaxDepth caps nested input elements; zero selects DefaultMaxDepth; negative values fail.
	MaxDepth int
	// AllowNonStrict opts into encoding/xml recovery and can change the interpreted tree.
	AllowNonStrict bool
	// ExpectedRoot requires an exact namespace/local-name match in Decode; zero disables the
	// check and Root only returns the actual name.
	ExpectedRoot xml.Name
	// CharsetReader is a borrowed conversion callback, invoked independently by validation and
	// target decoding. Nil uses the bounded built-in allowlist; custom work/output admission
	// is caller-owned.
	CharsetReader func(string, io.Reader) (io.Reader, error)
}

// EncodeOptions controls XML serialization.
type EncodeOptions struct {
	// MaxBytes caps returned encoded bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// Indent selects encoding/xml indentation, not canonical XML.
	Indent string
	// IncludeHeader prepends the XML declaration and counts it toward MaxBytes.
	// With a Charset, the declaration names that encoding.
	IncludeHeader bool
	// Charset selects the output encoding: UTF-8 when empty, or one of the
	// labels CharsetReader accepts (UTF-8, US-ASCII, ISO-8859-1 and
	// Windows-1252 with their aliases). A character the encoding cannot
	// represent fails as wire.ErrValidation; nothing is replaced or escaped.
	Charset string
}

// Decode parses one complete XML document into target.
func Decode(payload []byte, target any, options DecodeOptions) error {
	return DecodeReader(bytes.NewReader(payload), target, options)
}

// DecodeReader buffers borrowed reader input, checks one root/depth and optional
// ExpectedRoot, then independently decodes into a non-nil pointer. It never closes
// the reader; there is no cancellation API. Target mutation can precede failure.
// Custom charset/target callbacks must be bounded and input-stable across passes.
func DecodeReader(reader io.Reader, target any, options DecodeOptions) error {
	if err := validateTarget(target); err != nil {
		return &wire.Error{Kind: wire.ErrorKindTarget, Format: wire.FormatXML, Op: "decode", Err: err}
	}
	if err := validateDecodeOptions(options); err != nil {
		return err
	}
	payload, err := readBounded(reader, options.MaxBytes)
	if err != nil {
		return err
	}

	root, err := rootName(payload, options)
	if err != nil {
		return err
	}
	if options.ExpectedRoot != (xml.Name{}) && root != options.ExpectedRoot {
		return validationError("validate root", fmt.Errorf(
			"got {%s}%s, want {%s}%s",
			root.Space,
			root.Local,
			options.ExpectedRoot.Space,
			options.ExpectedRoot.Local,
		))
	}

	decoder := decoderFor(payload, options)
	if err := decoder.Decode(target); err != nil {
		return classifyDecodeError("decode", err)
	}

	return nil
}

// Root scans a bounded XML token stream and requires exactly one root element
// within the depth policy. It returns the resolved namespace/local name, ignoring
// ExpectedRoot. It is not schema/DTD validation and does not reject non-element
// character data outside that root. Charset callbacks own their conversion work.
func Root(payload []byte, options DecodeOptions) (xml.Name, error) {
	if err := validateDecodeOptions(options); err != nil {
		return xml.Name{}, err
	}
	bounded, err := readBounded(bytes.NewReader(payload), options.MaxBytes)
	if err != nil {
		return xml.Name{}, err
	}

	return rootName(bounded, options)
}

// Encode serializes value using encoding/xml's deterministic struct traversal.
func Encode(value any, options EncodeOptions) ([]byte, error) {
	if err := valuecheck.Validate(value); err != nil {
		return nil, validationError("encode", err)
	}
	encoding, err := outputEncoding(options.Charset)
	if err != nil {
		return nil, validationError("encode options", err)
	}
	if options.MaxBytes < 0 {
		return nil, validationError("encode options", errors.New("max bytes must not be negative"))
	}
	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	// MaxBytes caps the returned bytes. A supported single-byte character takes
	// at most three UTF-8 bytes (the Windows-1252 typographic range), so the
	// UTF-8 form is admitted up to three times the limit and the converted
	// output is checked against the limit itself.
	utf8Limit := maxBytes
	if encoding.encode != nil {
		utf8Limit = maxBytes * 3
		if maxBytes > math.MaxInt64/3 {
			utf8Limit = math.MaxInt64
		}
	}
	output, err := outputlimit.New(utf8Limit, DefaultMaxBytes)
	if err != nil {
		return nil, validationError("encode options", err)
	}
	if options.IncludeHeader {
		header := xml.Header
		if encoding.name != "UTF-8" {
			header = `<?xml version="1.0" encoding="` + encoding.name + `"?>` + "\n"
		}
		if _, err := output.Write([]byte(header)); err != nil {
			return nil, encodeSizeError(err)
		}
	}
	encoder := xml.NewEncoder(output)
	if options.Indent != "" {
		encoder.Indent("", options.Indent)
	}
	if err := encoder.Encode(value); err != nil {
		if errors.Is(err, outputlimit.ErrLimit) {
			return nil, encodeSizeError(err)
		}
		return nil, validationError("encode", err)
	}
	if encoding.encode == nil {
		return output.Bytes(), nil
	}
	encoded, err := encoding.encode(output.Bytes())
	if err != nil {
		return nil, validationError("encode", err)
	}
	if exceedsLimit(len(encoded), maxBytes) {
		return nil, encodeSizeError(outputlimit.ErrLimit)
	}
	return encoded, nil
}

// outputCharset names an output encoding and converts UTF-8 to it; a nil
// encode keeps UTF-8 bytes unchanged.
type outputCharset struct {
	name   string
	encode func([]byte) ([]byte, error)
}

func outputEncoding(label string) (outputCharset, error) {
	if len(label) > 64 {
		return outputCharset{}, errors.New("charset label exceeds size limit")
	}
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "", "utf-8", "utf8":
		return outputCharset{name: "UTF-8"}, nil
	case "us-ascii", "ascii":
		return outputCharset{name: "US-ASCII", encode: func(payload []byte) ([]byte, error) {
			return encodeSingleByte(payload, "US-ASCII", asciiByte)
		}}, nil
	case "iso-8859-1", "latin1", "latin-1":
		return outputCharset{name: "ISO-8859-1", encode: func(payload []byte) ([]byte, error) {
			return encodeSingleByte(payload, "ISO-8859-1", latin1Byte)
		}}, nil
	case "windows-1252", "cp1252":
		return outputCharset{name: "windows-1252", encode: func(payload []byte) ([]byte, error) {
			return encodeSingleByte(payload, "windows-1252", windows1252Byte)
		}}, nil
	default:
		return outputCharset{}, errors.New("unsupported charset")
	}
}

// encodeSingleByte maps each UTF-8 sequence to one output byte. Mappings read
// the sequence bytes so no rune is narrowed to a byte.
func encodeSingleByte(payload []byte, name string, mapping func(rune, []byte) (byte, bool)) ([]byte, error) {
	encoded := make([]byte, 0, len(payload))
	for index := 0; index < len(payload); {
		character, size := utf8.DecodeRune(payload[index:])
		value, ok := mapping(character, payload[index:index+size])
		if !ok {
			return nil, fmt.Errorf("character U+%04X is not representable in %s", character, name)
		}
		encoded = append(encoded, value)
		index += size
	}
	return encoded, nil
}

func asciiByte(_ rune, sequence []byte) (byte, bool) {
	return sequence[0], len(sequence) == 1 && sequence[0] < utf8.RuneSelf
}

// latin1Byte maps U+0000-U+00FF, whose UTF-8 form is one byte or a 0xC2/0xC3
// lead byte carrying the top two bits of the Latin-1 value.
func latin1Byte(character rune, sequence []byte) (byte, bool) {
	if len(sequence) == 1 {
		return sequence[0], sequence[0] < utf8.RuneSelf
	}
	if character > 0xff || len(sequence) != 2 {
		return 0, false
	}
	return (sequence[0]&0x03)<<6 | sequence[1]&0x3f, true
}

// windows1252Byte maps a character to Windows-1252, where 0x80-0x9F hold
// typographic characters instead of C1 controls and five positions are unused.
func windows1252Byte(character rune, sequence []byte) (byte, bool) {
	if character < 0x80 || character >= 0xa0 && character <= 0xff {
		return latin1Byte(character, sequence)
	}
	value, ok := windows1252Bytes[character]
	return value, ok
}

var windows1252Bytes = func() map[rune]byte {
	reverse := make(map[rune]byte, len(windows1252))
	for value, character := range windows1252 {
		reverse[character] = value
	}
	return reverse
}()

func encodeSizeError(err error) error {
	return &wire.Error{
		Kind:   wire.ErrorKindSizeLimit,
		Format: wire.FormatXML,
		Op:     "encode",
		Err:    errors.Join(ErrPayloadTooLarge, err),
	}
}

// EncodeWriter completes serialization before touching a borrowed writer, which
// it never closes. Encoding failure writes nothing; a later write error can leave
// a prefix visible and matches wire.ErrWrite. It is not atomic publication.
func EncodeWriter(writer io.Writer, value any, options EncodeOptions) error {
	payload, err := Encode(value, options)
	if err != nil {
		return err
	}
	if writer == nil {
		return validationError("write", errors.New("writer must not be nil"))
	}
	if _, err := io.Copy(writer, bytes.NewReader(payload)); err != nil {
		return &wire.Error{Kind: wire.ErrorKindWrite, Format: wire.FormatXML, Op: "write", Err: err}
	}

	return nil
}

// CharsetReader converts a deliberately limited set of common vendor
// encodings to UTF-8 with DefaultMaxBytes as the raw input limit.
// Labels are limited to 64 bytes before normalization. Conversion may produce
// up to three UTF-8 bytes per input byte. Caller reader code must return promptly.
func CharsetReader(label string, input io.Reader) (io.Reader, error) {
	return CharsetReaderWithLimit(label, input, 0)
}

// CharsetReaderWithLimit converts supported encodings with an inclusive raw
// input byte quota. Zero uses DefaultMaxBytes; negative limits are invalid.
// Unsupported or overlong labels are rejected without reading input.
// Errors expose categories only; wrapped reader causes are trusted diagnostics.
func CharsetReaderWithLimit(label string, input io.Reader, maxBytes int64) (io.Reader, error) {
	if maxBytes < 0 {
		return nil, charsetError{message: "invalid charset byte limit"}
	}
	if len(label) > 64 {
		return nil, charsetError{message: "charset label exceeds size limit"}
	}
	label = strings.ToLower(strings.TrimSpace(label))
	switch label {
	case "utf-8", "utf8", "us-ascii", "ascii", "iso-8859-1", "latin1", "latin-1", "windows-1252", "cp1252":
	default:
		return nil, charsetError{message: "unsupported charset"}
	}
	if input == nil {
		return nil, charsetError{message: "invalid charset reader"}
	}
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	readLimit := maxBytes
	if readLimit < math.MaxInt64 {
		readLimit++
	}
	payload, err := io.ReadAll(io.LimitReader(input, readLimit))
	if err != nil {
		return nil, charsetError{message: "charset input read failed", cause: err}
	}
	if int64(len(payload)) > maxBytes {
		return nil, charsetError{message: "charset input exceeds size limit", cause: ErrPayloadTooLarge}
	}

	switch label {
	case "utf-8", "utf8":
		if !utf8.Valid(payload) {
			return nil, charsetError{message: "invalid UTF-8"}
		}
		return bytes.NewReader(payload), nil
	case "us-ascii", "ascii":
		for _, value := range payload {
			if value > 0x7f {
				return nil, charsetError{message: "invalid US-ASCII"}
			}
		}
		return bytes.NewReader(payload), nil
	case "iso-8859-1", "latin1", "latin-1":
		return bytes.NewReader(singleByteToUTF8(payload, nil)), nil
	default:
		// Admission above leaves only the Windows-1252 aliases here.
		converted, err := windows1252ToUTF8(payload)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(converted), nil
	}
}

func decoderFor(payload []byte, options DecodeOptions) *xml.Decoder {
	decoder := xml.NewDecoder(bytes.NewReader(payload))
	decoder.Strict = !options.AllowNonStrict
	decoder.CharsetReader = options.CharsetReader
	if decoder.CharsetReader == nil {
		decoder.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
			return CharsetReaderWithLimit(label, input, options.MaxBytes)
		}
	}
	return decoder
}

func rootName(payload []byte, options DecodeOptions) (xml.Name, error) {
	decoder := decoderFor(payload, options)
	var root xml.Name
	depth := 0
	maxDepth := options.MaxDepth
	if maxDepth == 0 {
		maxDepth = DefaultMaxDepth
	}
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			if root == (xml.Name{}) {
				return xml.Name{}, parseError("read root", errors.New("document has no root element"))
			}
			return root, nil
		}
		if err != nil {
			return xml.Name{}, classifyDecodeError("read root", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				if root != (xml.Name{}) {
					return xml.Name{}, parseError("read root", errors.New("document has multiple root elements"))
				}
				root = typed.Name
			}
			depth++
			if depth > maxDepth {
				return xml.Name{}, &wire.Error{
					Kind:   wire.ErrorKindSizeLimit,
					Format: wire.FormatXML,
					Op:     "read root",
					Err:    ErrNestingTooDeep,
				}
			}
		case xml.EndElement:
			depth--
		}
	}
}

func validateDecodeOptions(options DecodeOptions) error {
	if options.MaxDepth < 0 {
		return validationError("decode options", errors.New("max depth must not be negative"))
	}
	return nil
}

func readBounded(reader io.Reader, configuredMax int64) ([]byte, error) {
	if reader == nil {
		return nil, validationError("read", errors.New("reader must not be nil"))
	}
	if configuredMax < 0 {
		return nil, validationError("read", errors.New("max bytes must not be negative"))
	}
	maxBytes := configuredMax
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
		return nil, &wire.Error{Kind: wire.ErrorKindSizeLimit, Format: wire.FormatXML, Op: "read", Err: ErrPayloadTooLarge}
	}
	return payload, nil
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

type charsetError struct {
	message string
	cause   error
}

func (e charsetError) Error() string {
	return e.message
}

func (e charsetError) Unwrap() error { return e.cause }

func classifyDecodeError(op string, err error) error {
	var syntaxError *xml.SyntaxError
	if errors.As(err, &syntaxError) {
		return parseError(op, err)
	}
	var encodingError charsetError
	if errors.As(err, &encodingError) {
		return parseError(op, err)
	}
	var numberError *strconv.NumError
	var unmarshalError xml.UnmarshalError
	if errors.As(err, &numberError) || errors.As(err, &unmarshalError) {
		return validationError(op, err)
	}
	return parseError(op, err)
}

func parseError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindParse, Format: wire.FormatXML, Op: op, Err: err}
}

func validationError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindValidation, Format: wire.FormatXML, Op: op, Err: err}
}

func singleByteToUTF8(payload []byte, replacements map[byte]rune) []byte {
	var output strings.Builder
	output.Grow(len(payload))
	for _, value := range payload {
		if replacement, ok := replacements[value]; ok {
			output.WriteRune(replacement)
		} else {
			output.WriteRune(rune(value))
		}
	}
	return []byte(output.String())
}

func windows1252ToUTF8(payload []byte) ([]byte, error) {
	for _, value := range payload {
		if isWindows1252Control(value) {
			if _, ok := windows1252[value]; !ok {
				return nil, charsetError{message: "undefined Windows-1252 code point"}
			}
		}
	}
	return singleByteToUTF8(payload, windows1252), nil
}

func exceedsLimit(length int, maximum int64) bool {
	return int64(length) > maximum
}

func isWindows1252Control(value byte) bool {
	return value >= 0x80 && value <= 0x9f
}

var windows1252 = map[byte]rune{
	0x80: '€', 0x82: '‚', 0x83: 'ƒ', 0x84: '„', 0x85: '…', 0x86: '†',
	0x87: '‡', 0x88: 'ˆ', 0x89: '‰', 0x8a: 'Š', 0x8b: '‹', 0x8c: 'Œ',
	0x8e: 'Ž', 0x91: '‘', 0x92: '’', 0x93: '“', 0x94: '”', 0x95: '•',
	0x96: '–', 0x97: '—', 0x98: '˜', 0x99: '™', 0x9a: 'š', 0x9b: '›',
	0x9c: 'œ', 0x9e: 'ž', 0x9f: 'Ÿ',
}
