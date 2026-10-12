package tomlwire

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/internal/outputlimit"
	"github.com/faustbrian/go-wire/v3/internal/valuecheck"
)

// DefaultMaxBytes is the default maximum TOML document size.
const DefaultMaxBytes int64 = 1 << 20

// ErrPayloadTooLarge identifies TOML documents over the configured limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

// DecodeOptions controls TOML decoding and schema strictness.
type DecodeOptions struct {
	// MaxBytes caps supplied input bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// DisallowUnknownFields rejects Undecoded keys after target decoding, which may already
	// have mutated target.
	DisallowUnknownFields bool
}

// EncodeOptions controls TOML serialization. Indent may contain only spaces
// and tabs; the default is two spaces.
type EncodeOptions struct {
	// MaxBytes caps returned encoded bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// Indent overrides the two-space encoder default using only spaces/tabs; empty keeps that
	// default.
	Indent string
}

// Decode parses one complete TOML document into target.
func Decode(payload []byte, target any, options DecodeOptions) error {
	return DecodeReader(bytes.NewReader(payload), target, options)
}

// DecodeReader buffers a borrowed reader through EOF or size refusal before target
// decoding. Reader is never closed and there is no context API. Target must be a
// non-nil pointer; codec/strictness failures can leave it mutated. Keep custom
// codecs bounded and do not reuse a failed target without resetting it.
func DecodeReader(reader io.Reader, target any, options DecodeOptions) error {
	if err := validateTarget(target); err != nil {
		return wrap(wire.ErrorKindTarget, "decode", err)
	}
	if options.MaxBytes < 0 {
		return wrap(wire.ErrorKindValidation, "decode options", errors.New("max bytes must not be negative"))
	}
	payload, err := readBounded(reader, options.MaxBytes)
	if err != nil {
		return err
	}
	metadata, err := toml.NewDecoder(bytes.NewReader(payload)).Decode(target)
	if err != nil {
		return classifyDecodeError(err)
	}
	if options.DisallowUnknownFields {
		if unknown := metadata.Undecoded(); len(unknown) > 0 {
			return wrap(wire.ErrorKindValidation, "decode", fmt.Errorf("unknown TOML keys: %v", unknown))
		}
	}
	return nil
}

// Encode borrows a stable acyclic document value and returns owned bytes with
// codec-sorted map keys. It preserves TOML-specific native values, not a universal
// cross-format model. The output quota does not bound custom marshaler work or
// intermediate allocations; failure returns no payload.
func Encode(value any, options EncodeOptions) ([]byte, error) {
	if err := valuecheck.Validate(value); err != nil {
		return nil, wrap(wire.ErrorKindEncode, "encode", err)
	}
	if strings.ContainsAny(options.Indent, "\r\n") || strings.Trim(options.Indent, " \t") != "" {
		return nil, wrap(wire.ErrorKindValidation, "encode options", errors.New("indent may contain only spaces and tabs"))
	}
	output, err := outputlimit.New(options.MaxBytes, DefaultMaxBytes)
	if err != nil {
		return nil, wrap(wire.ErrorKindValidation, "encode options", err)
	}
	encoder := toml.NewEncoder(output)
	if options.Indent != "" {
		encoder.Indent = options.Indent
	}
	if err := encoder.Encode(value); err != nil {
		if errors.Is(err, outputlimit.ErrLimit) {
			return nil, wrap(wire.ErrorKindSizeLimit, "encode", errors.Join(ErrPayloadTooLarge, err))
		}
		kind := wire.ErrorKindEncode
		message := err.Error()
		if strings.Contains(message, "unsupported type") || strings.Contains(message, "cannot encode") {
			kind = wire.ErrorKindUnsupported
		}
		return nil, wrap(kind, "encode", err)
	}
	return output.Bytes(), nil
}

// EncodeWriter completes serialization before touching a borrowed writer, which
// it never closes. Encoding failure writes nothing; a later write error can leave
// a prefix visible and matches wire.ErrWrite. It is not an atomic publication API.
func EncodeWriter(writer io.Writer, value any, options EncodeOptions) error {
	payload, err := Encode(value, options)
	if err != nil {
		return err
	}
	if writer == nil {
		return wrap(wire.ErrorKindValidation, "write", errors.New("writer must not be nil"))
	}
	if _, err := io.Copy(writer, bytes.NewReader(payload)); err != nil {
		return wrap(wire.ErrorKindWrite, "write", err)
	}
	return nil
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

func readBounded(reader io.Reader, configuredMax int64) ([]byte, error) {
	if reader == nil {
		return nil, wrap(wire.ErrorKindValidation, "read", errors.New("reader must not be nil"))
	}
	maxBytes := configuredMax
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	readLimit := maxBytes + 1
	if maxBytes == math.MaxInt64 {
		readLimit = maxBytes
	}
	payload, err := io.ReadAll(io.LimitReader(reader, readLimit))
	if err != nil {
		return nil, wrap(wire.ErrorKindParse, "read", err)
	}
	if int64(len(payload)) > maxBytes {
		return nil, wrap(wire.ErrorKindSizeLimit, "read", ErrPayloadTooLarge)
	}
	return payload, nil
}

func classifyDecodeError(err error) error {
	var parseError toml.ParseError
	if errors.As(err, &parseError) {
		message := parseError.Message
		if strings.Contains(message, "out of range") ||
			strings.Contains(message, "incompatible types") ||
			strings.Contains(message, "type mismatch") {
			return wrap(wire.ErrorKindValidation, "decode", err)
		}
		return wrap(wire.ErrorKindParse, "decode", err)
	}
	return wrap(wire.ErrorKindValidation, "decode", err)
}

func wrap(kind wire.ErrorKind, op string, err error) error {
	return &wire.Error{Kind: kind, Format: wire.FormatTOML, Op: op, Err: err}
}
