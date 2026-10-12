package wire

import (
	"errors"
	"fmt"
)

// ErrorKind classifies a failure independently of its underlying cause.
type ErrorKind string

const (
	// ErrorKindParse classifies malformed input and reader failures.
	ErrorKindParse ErrorKind = "parse"
	// ErrorKindValidation classifies option, value and semantic validation failures.
	ErrorKindValidation ErrorKind = "validation"
	// ErrorKindUnsupported classifies unsupported formats or codec features.
	ErrorKindUnsupported ErrorKind = "unsupported"
	// ErrorKindEnvelope classifies invalid SOAP envelope/body/fault structure.
	ErrorKindEnvelope ErrorKind = "envelope"
	// ErrorKindFault classifies a valid SOAP fault response.
	ErrorKindFault ErrorKind = "fault"
	// ErrorKindWrite classifies a destination failure after successful serialization.
	ErrorKindWrite ErrorKind = "write"
	// ErrorKindSizeLimit classifies a configured byte or structural budget refusal.
	ErrorKindSizeLimit ErrorKind = "size-limit"
	// ErrorKindTarget classifies nil, typed-nil or non-pointer decode targets.
	ErrorKindTarget ErrorKind = "target"
	// ErrorKindEncode classifies codec serialization failures where the format uses this
	// category.
	ErrorKindEncode ErrorKind = "encode"
)

var (
	// ErrParse is the errors.Is category for malformed input or read failures.
	ErrParse = errors.New("parse failure")
	// ErrValidation is the errors.Is category for invalid options, assignments or values.
	ErrValidation = errors.New("validation failure")
	// ErrUnsupportedFormat is the errors.Is category for unsupported formats or features.
	ErrUnsupportedFormat = errors.New("unsupported format")
	// ErrEnvelope is the errors.Is category for invalid SOAP structure.
	ErrEnvelope = errors.New("envelope failure")
	// ErrSOAPFault is the errors.Is category for a valid SOAP fault response.
	ErrSOAPFault = errors.New("SOAP fault")
	// ErrWrite is the errors.Is category for destination write failures.
	ErrWrite = errors.New("write failure")
	// ErrSizeLimit is the errors.Is category for byte or structural budget refusal.
	ErrSizeLimit = errors.New("size limit exceeded")
	// ErrTarget is the errors.Is category for invalid decode targets.
	ErrTarget = errors.New("invalid target")
	// ErrEncode is the errors.Is category for codec serialization failure.
	ErrEncode = errors.New("encode failure")
)

// Error describes a wire-format failure using stable categories. Causes and
// structured fields remain inspectable, not sanitized; redact before disclosure.
// Error/Is/Unwrap require a non-nil receiver. There is no automatic cause copy.
type Error struct {
	// Kind selects the stable public category used by Error and Is.
	Kind ErrorKind
	// Format identifies the codec; it is inspectable metadata, not rendered by Error.
	Format Format
	// Op contains diagnostic operation text; caller-supplied values are not sanitized.
	Op string
	// Err retains the original cause for Unwrap/Is/As and can contain sensitive input.
	Err error
}

// Error returns a stable classification without rendering diagnostic fields
// or the underlying cause, which may contain untrusted wire data.
func (e *Error) Error() string {
	if kindErr := sentinelForKind(e.Kind); kindErr != nil {
		return "wire: " + kindErr.Error()
	}
	return "wire: failure"
}

// Unwrap returns the underlying format-specific failure, if any.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether target is the sentinel associated with the error kind or
// matches the underlying cause.
func (e *Error) Is(target error) bool {
	return target == sentinelForKind(e.Kind) || errors.Is(e.Err, target)
}

func sentinelForKind(kind ErrorKind) error {
	switch kind {
	case ErrorKindParse:
		return ErrParse
	case ErrorKindValidation:
		return ErrValidation
	case ErrorKindUnsupported:
		return ErrUnsupportedFormat
	case ErrorKindEnvelope:
		return ErrEnvelope
	case ErrorKindFault:
		return ErrSOAPFault
	case ErrorKindWrite:
		return ErrWrite
	case ErrorKindSizeLimit:
		return ErrSizeLimit
	case ErrorKindTarget:
		return ErrTarget
	case ErrorKindEncode:
		return ErrEncode
	default:
		return nil
	}
}

func newError(kind ErrorKind, format Format, op string, err error) error {
	return &Error{Kind: kind, Format: format, Op: op, Err: err}
}

func unsupportedError(op string, payload []byte) error {
	var cause error
	if len(payload) > 0 {
		cause = fmt.Errorf("unrecognized leading byte %q", payload[0])
	}

	return newError(ErrorKindUnsupported, "", op, cause)
}
