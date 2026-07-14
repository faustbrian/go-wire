package jsonwire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/faustbrian/go-wire"
)

// DefaultMaxBytes is the default maximum payload size accepted by decoders.
const DefaultMaxBytes int64 = 1 << 20

// ErrPayloadTooLarge identifies payloads that exceed the configured limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

// DecodeOptions controls JSON decoding behavior.
type DecodeOptions struct {
	MaxBytes              int64
	DisallowUnknownFields bool
}

// EncodeOptions controls JSON encoding behavior.
type EncodeOptions struct {
	Indent              string
	DisableHTMLEscaping bool
}

// NormalizeOptions controls JSON normalization behavior.
type NormalizeOptions struct {
	MaxBytes int64
}

// Decode decodes exactly one JSON value from payload into target.
func Decode(payload []byte, target any, options DecodeOptions) error {
	return DecodeReader(bytes.NewReader(payload), target, options)
}

// DecodeReader reads a bounded payload and decodes exactly one JSON value.
func DecodeReader(reader io.Reader, target any, options DecodeOptions) error {
	if err := validateTarget(target); err != nil {
		return validationError("decode", err)
	}

	payload, err := readBounded(reader, options.MaxBytes)
	if err != nil {
		return err
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

// Encode serializes value deterministically. Map keys are ordered according to
// encoding/json's stable key ordering.
func Encode(value any, options EncodeOptions) ([]byte, error) {
	if !options.DisableHTMLEscaping {
		var (
			payload []byte
			err     error
		)
		if options.Indent == "" {
			payload, err = json.Marshal(value)
		} else {
			payload, err = json.MarshalIndent(value, "", options.Indent)
		}
		if err != nil {
			return nil, validationError("encode", err)
		}

		return payload, nil
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	if options.Indent != "" {
		encoder.SetIndent("", options.Indent)
	}
	if err := encoder.Encode(value); err != nil {
		return nil, validationError("encode", err)
	}

	return bytes.TrimSuffix(output.Bytes(), []byte{'\n'}), nil
}

// EncodeWriter serializes value deterministically and writes the complete JSON
// document to writer.
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
// precision or representation.
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

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, decodeErrorWithOp("normalize", err)
	}
	if err := requireEOFWithOp(decoder, "normalize"); err != nil {
		return nil, err
	}

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
		return nil, validationError("read", ErrPayloadTooLarge)
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
