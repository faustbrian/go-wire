package bsonwire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/internal/outputlimit"
	"github.com/faustbrian/go-wire/v3/internal/valuecheck"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Official BSON data types are aliases rather than a second local type system;
// the pinned driver retains their methods, representation and ownership semantics.
type (
	// A is the official driver array type; it is not a top-level BSON document.
	A = bson.A
	// D is an ordered driver document; retain field order and explicit duplicate opt-in.
	D = bson.D
	// E is one key/value element of D.
	E = bson.E
	// M is a driver map document with unspecified encoding order.
	M = bson.M
	// Raw is the driver byte-backed document type; alias conversion does not copy its bytes.
	Raw = bson.Raw
	// RawArray is the driver byte-backed array type, not a top-level document.
	RawArray = bson.RawArray
	// RawValue is a driver type tag and encoded value view; underlying bytes can alias their
	// owner.
	RawValue = bson.RawValue
	// ObjectID is the official driver 12-byte object identifier.
	ObjectID = bson.ObjectID
	// DateTime is the driver signed millisecond Unix-time representation.
	DateTime = bson.DateTime
	// Decimal128 is the official driver BSON decimal128 representation.
	Decimal128 = bson.Decimal128
	// Binary is the driver binary value including subtype and mutable byte data.
	Binary = bson.Binary
	// Regex is the driver pattern/options description, not a compiled regular expression.
	Regex = bson.Regex
	// Timestamp is the driver BSON timestamp pair, not a DateTime alias.
	Timestamp = bson.Timestamp
	// JavaScript is BSON JavaScript source data; this package does not execute it.
	JavaScript = bson.JavaScript
	// CodeWithScope is driver JavaScript source and scope data; Encode/Decode validate
	// scope structure without executing code.
	CodeWithScope = bson.CodeWithScope
	// MinKey is the driver BSON minimum-key marker.
	MinKey = bson.MinKey
	// MaxKey is the driver BSON maximum-key marker.
	MaxKey = bson.MaxKey
	// Undefined is the driver BSON undefined marker.
	Undefined = bson.Undefined
	// DBPointer is the driver legacy database-pointer value.
	DBPointer = bson.DBPointer
	// Symbol is the driver BSON symbol string.
	Symbol = bson.Symbol
)

// DefaultMaxBytes is the default maximum BSON document size.
const DefaultMaxBytes int64 = 1 << 20

// DefaultMaxNestedLevels is the maximum raw-validation nesting below the root.
// Documents, arrays, and CodeWithScope scope documents each add one level.
// Encoding checks this policy after codec work, not before its allocation.
const DefaultMaxNestedLevels = 100

// ErrPayloadTooLarge identifies BSON documents over the configured byte limit.
var ErrPayloadTooLarge = errors.New("payload exceeds size limit")

var errDuplicateKey = errors.New("duplicate BSON key")

var errNestingLimit = errors.New("BSON nesting limit exceeded")

// DecodeOptions controls BSON decoding and explicit interoperability choices.
type DecodeOptions struct {
	// MaxBytes caps supplied input bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// AllowDuplicateKeys opts into duplicate document names; structural validation remains
	// active, and preservation depends on the driver target.
	AllowDuplicateKeys bool
	// ObjectIDAsHexString permits the driver to decode ObjectID values into string targets.
	ObjectIDAsHexString bool
	// AllowTruncatingDoubles permits the driver to truncate fractional double values when
	// assigning integer targets.
	AllowTruncatingDoubles bool
}

// EncodeOptions controls BSON document encoding.
type EncodeOptions struct {
	// MaxBytes caps returned encoded bytes; zero selects DefaultMaxBytes; negative values fail.
	MaxBytes int64
	// AllowDuplicateKeys opts into duplicate document names; structural validation remains
	// active, and preservation depends on the driver target.
	AllowDuplicateKeys bool
	// MinimizeIntegerWidth enables driver integer-width minimization, which changes BSON
	// representation.
	MinimizeIntegerWidth bool
	// UseJSONStructTags uses JSON tags for fields lacking BSON tags, according to the driver.
	UseJSONStructTags bool
}

// Decode parses exactly one complete BSON document into target.
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
	raw, err := validateDocument(payload, options.AllowDuplicateKeys)
	if err != nil {
		if errors.Is(err, errNestingLimit) {
			return wrap(wire.ErrorKindSizeLimit, "decode", err)
		}
		return wrap(wire.ErrorKindParse, "decode", err)
	}
	decoder := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(raw)))
	if options.ObjectIDAsHexString {
		decoder.ObjectIDAsHexString()
	}
	if options.AllowTruncatingDoubles {
		decoder.AllowTruncatingDoubles()
	}
	if err := decoder.Decode(target); err != nil {
		return wrap(wire.ErrorKindValidation, "decode", err)
	}
	return nil
}

// Encode serializes a BSON document. Struct and D ordering is stable; M map
// ordering follows Go map iteration and is intentionally not guaranteed.
// Input/custom codecs are borrowed and must remain stable. The output buffer
// limit does not bound driver buffering; raw depth/duplicate validation follows
// encoding. Failure returns no payload, including post-encoding validation.
func Encode(value any, options EncodeOptions) ([]byte, error) {
	if err := valuecheck.Validate(value); err != nil {
		return nil, wrap(wire.ErrorKindEncode, "encode", err)
	}
	if !isDocumentValue(value) {
		return nil, wrap(wire.ErrorKindUnsupported, "encode", errors.New("BSON requires a top-level document"))
	}
	output, err := outputlimit.New(options.MaxBytes, DefaultMaxBytes)
	if err != nil {
		return nil, wrap(wire.ErrorKindValidation, "encode options", err)
	}
	encoder := bson.NewEncoder(bson.NewDocumentWriter(output))
	if options.MinimizeIntegerWidth {
		encoder.IntMinSize()
	}
	if options.UseJSONStructTags {
		encoder.UseJSONStructTags()
	}
	if err := encoder.Encode(value); err != nil {
		if errors.Is(err, outputlimit.ErrLimit) {
			return nil, wrap(wire.ErrorKindSizeLimit, "encode", errors.Join(ErrPayloadTooLarge, err))
		}
		return nil, wrap(wire.ErrorKindEncode, "encode", err)
	}
	if _, err := validateDocument(output.Bytes(), options.AllowDuplicateKeys); err != nil {
		if errors.Is(err, errNestingLimit) {
			return nil, wrap(wire.ErrorKindSizeLimit, "encode", err)
		}
		return nil, wrap(wire.ErrorKindValidation, "encode", err)
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

func isDocumentValue(value any) bool {
	if value == nil {
		return false
	}
	if _, ok := value.(bson.Marshaler); ok {
		return true
	}
	typeOf := reflect.TypeOf(value)
	for typeOf.Kind() == reflect.Pointer {
		typeOf = typeOf.Elem()
	}
	if typeOf == reflect.TypeFor[bson.D]() || typeOf == reflect.TypeFor[bson.Raw]() {
		return true
	}
	return typeOf.Kind() == reflect.Struct || typeOf.Kind() == reflect.Map
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

func validateDocument(payload []byte, allowDuplicates bool) (bson.Raw, error) {
	raw := bson.Raw(payload)
	frame, err := newStructureFrame(raw, false, 0, allowDuplicates)
	if err != nil {
		return nil, err
	}
	stack := []structureFrame{frame}
	for len(stack) != 0 {
		current := &stack[len(stack)-1]
		if current.index == current.length() {
			stack = stack[:len(stack)-1]
			continue
		}
		value := current.value()
		current.index++
		var child bson.Raw
		isArray := false
		switch value.Type {
		case bson.TypeEmbeddedDocument:
			// Frame admission already validated this value's encoded length.
			// Its own document structure is checked on child-frame entry.
			child = bson.Raw(value.Value)
		case bson.TypeArray:
			child, isArray = bson.Raw(value.Value), true
		case bson.TypeCodeWithScope:
			child, err = scopeDocument(value)
			if err != nil {
				return nil, err
			}
		default:
			continue
		}
		frame, err := newStructureFrame(child, isArray, current.depth+1, allowDuplicates)
		if err != nil {
			return nil, err
		}
		stack = append(stack, frame)
	}
	return raw, nil
}

type structureFrame struct {
	elements []bson.RawElement
	values   []bson.RawValue
	index    int
	depth    int
}

func (f *structureFrame) length() int {
	if f.values != nil {
		return len(f.values)
	}
	return len(f.elements)
}

func (f *structureFrame) value() bson.RawValue {
	if f.values != nil {
		return f.values[f.index]
	}
	return f.elements[f.index].Value()
}

func newStructureFrame(raw bson.Raw, isArray bool, depth int, allowDuplicates bool) (structureFrame, error) {
	if depth > DefaultMaxNestedLevels {
		return structureFrame{}, errNestingLimit
	}
	payload := []byte(raw)
	if len(payload) < 5 {
		return structureFrame{}, errors.New("BSON document is shorter than its minimum length")
	}
	declared := int64(binary.LittleEndian.Uint32(payload[:4]))
	if declared != int64(len(payload)) {
		return structureFrame{}, errors.New("BSON length prefix does not match payload length")
	}
	if isArray {
		array := bson.RawArray(raw)
		if err := array.Validate(); err != nil {
			return structureFrame{}, err
		}
		values, err := array.Values()
		return structureFrame{values: values, depth: depth}, err
	}
	elements, err := raw.Elements()
	if err != nil {
		return structureFrame{}, err
	}
	// Elements validates values, but not the document's final terminator.
	// Driver validation remains shallow; each child is checked on entry by
	// the iterative traversal before any decoder callback is invoked.
	if err := raw.Validate(); err != nil {
		return structureFrame{}, err
	}
	if !allowDuplicates {
		seen := make(map[string]struct{}, len(elements))
		for _, element := range elements {
			key := element.Key()
			if _, exists := seen[key]; exists {
				return structureFrame{}, errors.Join(errDuplicateKey, errors.New(key))
			}
			seen[key] = struct{}{}
		}
	}
	return structureFrame{elements: elements, depth: depth}, nil
}

func scopeDocument(value bson.RawValue) (bson.Raw, error) {
	data := value.Value
	if len(data) < 14 {
		return nil, errors.New("invalid BSON code with scope")
	}
	codeLength := int64(binary.LittleEndian.Uint32(data[4:8]))
	if codeLength < 1 || codeLength > int64(len(data)-13) || data[8+int(codeLength)-1] != 0 {
		return nil, errors.New("invalid BSON scope code string")
	}
	_, scope, ok := value.CodeWithScopeOK()
	if !ok || len(scope) != len(data)-8-int(codeLength) {
		return nil, errors.New("invalid BSON scope document boundary")
	}
	return scope, nil
}

func wrap(kind wire.ErrorKind, op string, err error) error {
	return &wire.Error{Kind: kind, Format: wire.FormatBSON, Op: op, Err: err}
}
