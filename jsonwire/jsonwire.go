package jsonwire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"unicode/utf8"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/internal/outputlimit"
	"github.com/faustbrian/go-wire/v3/internal/valuecheck"
)

// DefaultMaxBytes is the default maximum payload size accepted by decoders.
const DefaultMaxBytes int64 = 1 << 20

// ErrPayloadTooLarge identifies payloads that exceed the configured limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

// maxNestedLevelsLimit is encoding/json's own nesting cap; a larger
// MaxNestedLevels could never be reached by a decodable payload.
const maxNestedLevelsLimit = 10_000

// DecodeOptions controls JSON decoding behavior.
type DecodeOptions struct {
	// MaxBytes caps supplied input bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// DisallowUnknownFields rejects unknown object fields when decoding into structs.
	DisallowUnknownFields bool
	// DisallowDuplicateNames rejects recursively repeated object names before target decoding;
	// default retains encoding/json replacement/merge behavior.
	DisallowDuplicateNames bool
	// MaxNestedLevels caps nested arrays and objects before target decoding; a
	// root array or object is level one. Zero retains encoding/json's own 10,000
	// level limit; negative values and values above 10,000 fail. Refusals match
	// wire.ErrSizeLimit.
	MaxNestedLevels int
	// ReplaceInvalidUTF8 admits invalid UTF-8 inside strings and replaces each
	// invalid byte with U+FFFD in decoded strings and names, as encoding/json
	// does, instead of rejecting the payload. json.RawMessage and json.Unmarshaler
	// targets receive the original invalid bytes. Invalid bytes outside strings
	// remain malformed JSON. Use it only for third-party input whose producer
	// cannot be held to RFC 8259.
	ReplaceInvalidUTF8 bool
}

// EncodeOptions controls JSON encoding behavior.
type EncodeOptions struct {
	// MaxBytes caps returned encoded bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// Indent selects encoding/json indentation; empty produces compact output.
	Indent string
	// DisableHTMLEscaping disables escaping of HTML-sensitive characters; it does not make
	// output safe for HTML embedding.
	DisableHTMLEscaping bool
}

// NormalizeOptions controls JSON normalization behavior.
type NormalizeOptions struct {
	// MaxBytes caps supplied input bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
}

// Decode borrows payload until return and by default decodes exactly one
// UTF-8 JSON value into a non-nil pointer. Untyped numbers follow encoding/json
// float64 mapping; choose typed numeric destinations when that representation
// is insufficient. Duplicate rejection, nesting budgets and invalid UTF-8
// replacement are opt-in and failed targets are indeterminate.
func Decode(payload []byte, target any, options DecodeOptions) error {
	return DecodeReader(bytes.NewReader(payload), target, options)
}

// DecodeReader buffers borrowed input through EOF or a size refusal and by
// default decodes exactly one UTF-8 JSON value into a non-nil pointer. It never
// closes reader. Duplicate and nesting preflight, when enabled, precedes
// assignment; other errors, including trailing values, can leave target
// mutated. Blocking/custom codec work is not cancellable through this API.
// Size and nesting refusals match wire.ErrSizeLimit.
func DecodeReader(reader io.Reader, target any, options DecodeOptions) error {
	if err := validateTarget(target); err != nil {
		return &wire.Error{Kind: wire.ErrorKindTarget, Format: wire.FormatJSON, Op: "decode", Err: err}
	}

	if options.MaxNestedLevels < 0 || options.MaxNestedLevels > maxNestedLevelsLimit {
		return validationError("decode", errors.New("max nested levels must be between 0 and 10000"))
	}
	payload, err := readBounded(reader, options.MaxBytes)
	if err != nil {
		return err
	}
	if !options.ReplaceInvalidUTF8 && !utf8.Valid(payload) {
		return parseError("decode", errors.New("JSON text is not valid UTF-8"))
	}
	if options.DisallowDuplicateNames || options.MaxNestedLevels > 0 {
		if err := validateStructure(payload, options); err != nil {
			return err
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	if options.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(target); err != nil {
		return decodeError(err)
	}
	if err := requireEOF(decoder); err != nil {
		return err
	}

	return nil
}

// validateStructure walks the payload once before assignment, refusing
// repeated object names and nesting beyond the configured level budget.
func validateStructure(payload []byte, options DecodeOptions) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	// Numbers are only skipped here; json.Number avoids refusing a value such
	// as 1e400 that the target itself may accept.
	decoder.UseNumber()
	if err := validateValue(decoder, options, 0); err != nil {
		return err
	}
	return requireEOF(decoder)
}

func validateValue(decoder *json.Decoder, options DecodeOptions, level int) error {
	token, err := decoder.Token()
	if err != nil {
		return decodeError(err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	level++
	if options.MaxNestedLevels > 0 && level > options.MaxNestedLevels {
		return &wire.Error{
			Kind: wire.ErrorKindSizeLimit, Format: wire.FormatJSON, Op: "decode",
			Err: fmt.Errorf("JSON nesting exceeds %d levels", options.MaxNestedLevels),
		}
	}
	switch delimiter {
	case '{':
		var names map[string]struct{}
		if options.DisallowDuplicateNames {
			names = make(map[string]struct{})
		}
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return decodeError(err)
			}
			name := token.(string) // encoding/json guarantees object names are strings.
			if options.DisallowDuplicateNames {
				if _, exists := names[name]; exists {
					return validationError("decode", errors.New("duplicate JSON object name"))
				}
				names[name] = struct{}{}
			}
			if err := validateValue(decoder, options, level); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return decodeError(err)
	}
	// A delimiter returned at the beginning of a JSON value can only be '[' or
	// '{'; the object case is handled above.
	for decoder.More() {
		if err := validateValue(decoder, options, level); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return decodeError(err)
}

// Encode borrows a stable acyclic value, orders built-in map keys using encoding/json
// and returns owned bytes without the encoder newline. MaxBytes bounds final
// output, not graph preflight or custom marshaler work. Errors return no payload;
// caller marshalers can change determinism and must terminate.
func Encode(value any, options EncodeOptions) ([]byte, error) {
	if err := valuecheck.Validate(value); err != nil {
		return nil, validationError("encode", err)
	}
	maxBytes := options.MaxBytes
	if maxBytes < 0 {
		return nil, validationError("encode options", errors.New("max bytes must not be negative"))
	}
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	bufferLimit := maxBytes
	if bufferLimit < math.MaxInt64 {
		bufferLimit++ // json.Encoder appends one newline that Encode removes.
	}
	output, _ := outputlimit.New(bufferLimit, bufferLimit)
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(!options.DisableHTMLEscaping)
	if options.Indent != "" {
		encoder.SetIndent("", options.Indent)
	}
	if err := encoder.Encode(value); err != nil {
		if errors.Is(err, outputlimit.ErrLimit) {
			return nil, encodeSizeError(err)
		}
		return nil, validationError("encode", err)
	}
	payload := bytes.TrimSuffix(output.Bytes(), []byte{'\n'})
	return payload, nil
}

func encodeSizeError(err error) error {
	return &wire.Error{
		Kind:   wire.ErrorKindSizeLimit,
		Format: wire.FormatJSON,
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
		return &wire.Error{Kind: wire.ErrorKindWrite, Format: wire.FormatJSON, Op: "write", Err: err}
	}

	return nil
}

// Normalize strips a UTF-8 BOM and insignificant whitespace, validates that
// the payload contains exactly one value, and emits compact, key-ordered JSON.
// Number lexemes are retained so normalization does not silently change vendor
// precision or representation. Duplicate names follow encoding/json replacement
// behavior; no RFC 8785/signing profile is claimed. MaxBytes only admits input:
// re-encoding uses the default 1 MiB output limit, even when input admission is raised.
// Source bytes are not mutated; failure returns no payload.
func Normalize(payload []byte, options NormalizeOptions) ([]byte, error) {
	if options.MaxBytes < 0 {
		return nil, validationError("normalize", fmt.Errorf("max bytes must not be negative"))
	}

	maxBytes := options.MaxBytes
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	if int64(len(payload)) > maxBytes {
		return nil, validationError("normalize", ErrPayloadTooLarge)
	}

	payload = bytes.TrimSpace(payload)
	payload = bytes.TrimPrefix(payload, []byte{0xef, 0xbb, 0xbf})
	payload = bytes.TrimSpace(payload)
	if !utf8.Valid(payload) {
		return nil, parseError("normalize", errors.New("JSON text is not valid UTF-8"))
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, decodeErrorWithOp("normalize", err)
	}
	if err := requireEOFWithOp(decoder, "normalize"); err != nil {
		return nil, err
	}

	// Normalization has its own input allowance but retains Encode defaults for
	// output; forwarding input MaxBytes would change accepted output behavior.
	return Encode(value, EncodeOptions{})
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
	if int64(len(payload)) > maxBytes {
		return nil, &wire.Error{Kind: wire.ErrorKindSizeLimit, Format: wire.FormatJSON, Op: "read", Err: ErrPayloadTooLarge}
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

func requireEOF(decoder *json.Decoder) error {
	return requireEOFWithOp(decoder, "decode")
}

func requireEOFWithOp(decoder *json.Decoder, op string) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		err = errors.New("multiple JSON values")
	}

	return parseError(op, err)
}

func decodeError(err error) error {
	if err == nil {
		return nil
	}

	return decodeErrorWithOp("decode", err)
}

func decodeErrorWithOp(op string, err error) error {
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) || bytes.Contains([]byte(err.Error()), []byte("unknown field")) {
		return validationError(op, err)
	}

	return parseError(op, err)
}

func parseError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindParse, Format: wire.FormatJSON, Op: op, Err: err}
}

func validationError(op string, err error) error {
	return &wire.Error{Kind: wire.ErrorKindValidation, Format: wire.FormatJSON, Op: op, Err: err}
}
