package msgpackwire

import (
	"encoding"
	"reflect"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// Custom decoding is caller-owned. Preflight must never execute a codec merely
// to predict the key it will publish during the real destination decode.
func hasCustomProjection(target reflect.Type) bool {
	for _, contract := range []reflect.Type{
		reflect.TypeFor[msgpack.CustomDecoder](), reflect.TypeFor[msgpack.Unmarshaler](),
		reflect.TypeFor[encoding.BinaryUnmarshaler](), reflect.TypeFor[encoding.TextUnmarshaler](),
	} {
		if target.Implements(contract) || target.Kind() != reflect.Pointer && reflect.PointerTo(target).Implements(contract) {
			return true
		}
	}
	return false
}

func rejectProjectedKeys(source any, target reflect.Type, loose bool, budget *admissionBudget) error {
	if hasCustomProjection(target) {
		return nil
	}
	for target.Kind() == reflect.Pointer {
		target = target.Elem()
		if hasCustomProjection(target) {
			return nil
		}
	}
	switch target.Kind() {
	case reflect.Map, reflect.Interface, reflect.Struct:
		object, ok := source.(numericMap)
		if !ok {
			break
		}
		keyType, valueType := reflect.TypeFor[string](), reflect.TypeFor[any]()
		var fields []numericStructField
		if target.Kind() == reflect.Struct {
			var supported bool
			fields, supported = projectionFields(target)
			if !supported {
				return nil
			}
		}
		if target.Kind() == reflect.Map {
			keyType, valueType = target.Key(), target.Elem()
		}
		if err := budget.reserveArrayKeys(keyType, len(object)); err != nil {
			return err
		}
		var keys []any
		for _, entry := range object {
			key, supported := projectedKey(entry.key, keyType, loose)
			if target.Kind() == reflect.Struct {
				matched := false
				for _, field := range fields {
					if key == field.name {
						valueType, matched = field.target, true
						break
					}
				}
				if !matched {
					continue // Unknown fields have no destination projection.
				}
			}
			if supported {
				for _, previous := range keys {
					if equalProjectedKeys(previous, key) {
						return errDuplicateKey
					}
				}
				keys = append(keys, key)
			}
			if err := rejectProjectedKeys(entry.value, valueType, loose, budget); err != nil {
				return err
			}
		}
		return nil
	}
	if items, ok := source.([]any); ok {
		switch target.Kind() {
		case reflect.Array, reflect.Slice, reflect.Interface:
			element := reflect.TypeFor[any]()
			if target.Kind() != reflect.Interface {
				element = target.Elem()
			}
			for _, item := range items {
				if err := rejectProjectedKeys(item, element, loose, budget); err != nil {
					return err
				}
			}
		case reflect.Struct:
			fields, supported := projectionFields(target)
			if !supported {
				return nil
			}
			for index, field := range fields {
				if index < len(items) {
					if err := rejectProjectedKeys(items[index], field.target, loose, budget); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Alias and auto-inlined field routing are deliberately not predicted here.
// The original raw-key and numeric-fit checks still apply to these targets.
func projectionFields(target reflect.Type) ([]numericStructField, bool) {
	fields := make([]numericStructField, 0, target.NumField())
	names := make(map[string]bool)
	for index := range target.NumField() {
		field := target.Field(index)
		tag := strings.Split(field.Tag.Get("msgpack"), ",")
		if tag[0] == "-" || field.Name == "_msgpack" {
			continue
		}
		if field.Anonymous && !hasTagOption(tag[1:], "noinline") {
			return nil, false
		}
		for _, option := range tag[1:] {
			if strings.HasPrefix(option, "alias:") || option == "intern" {
				return nil, false
			}
		}
		if !field.IsExported() {
			continue
		}
		name := tag[0]
		if name == "" {
			name = field.Name
		}
		if names[name] {
			return nil, false
		}
		names[name] = true
		fields = append(fields, numericStructField{name: name, target: field.Type})
	}
	return fields, true
}

func equalProjectedKeys(left, right any) bool {
	if items, ok := left.([]any); ok {
		other, ok := right.([]any)
		if !ok || len(items) != len(other) {
			return false
		}
		for index, item := range items {
			if !equalProjectedKeys(item, other[index]) {
				return false
			}
		}
		return true
	}
	// Non-composite projections are comparable. Go equality, unlike DeepEqual,
	// preserves pointer identity and the fact that NaN is not equal to itself.
	return left == right
}

// projectedKey describes only built-in destination equality, not a second
// decoder. Unsupported shapes retain the driver's existing validation path.
func projectedKey(source any, target reflect.Type, loose bool) (any, bool) {
	if hasCustomProjection(target) {
		return nil, false
	}
	if target.Kind() == reflect.Interface {
		if target.NumMethod() != 0 {
			return nil, false // Non-empty interfaces may publish driver-owned identities.
		}
		if source == nil {
			return nil, true
		}
		value := reflect.ValueOf(source)
		if loose {
			switch value.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32:
				return value.Int(), true
			case reflect.Uint8, reflect.Uint16, reflect.Uint32:
				return value.Uint(), true
			case reflect.Float32:
				return value.Convert(reflect.TypeFor[float64]()).Interface(), true
			case reflect.Slice:
				if bytes, ok := source.([]byte); ok {
					return string(bytes), true
				}
			}
		}
		return source, value.Comparable()
	}
	value := reflect.New(target).Elem()
	if source == nil && target.Kind() != reflect.Array && target.Kind() != reflect.Struct {
		return value.Interface(), true
	}
	switch target.Kind() {
	case reflect.String:
		switch source := source.(type) {
		case string:
			value.SetString(source)
		case []byte:
			value.SetString(string(source))
		default:
			return nil, false
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer, ok := integerValue(source)
		if !ok || value.OverflowInt(integer) {
			return nil, false
		}
		value.SetInt(integer)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		integer, ok := unsignedValue(source)
		if !ok || value.OverflowUint(integer) {
			return nil, false
		}
		value.SetUint(integer)
	case reflect.Float32, reflect.Float64:
		number, ok := floatValue(source)
		if !ok {
			integer, integerOK := integerValue(source)
			if !integerOK {
				unsigned, unsignedOK := unsignedValue(source)
				if !unsignedOK {
					return nil, false
				}
				integer = int64(unsigned)
			}
			number = float64(integer)
		} else if target.Kind() == reflect.Float32 && reflect.TypeOf(source).Kind() == reflect.Float64 {
			return nil, false // The driver does not decode a double into float32.
		}
		value.SetFloat(number)
	case reflect.Bool:
		boolean, ok := source.(bool)
		if !ok {
			return nil, false
		}
		value.SetBool(boolean)
	case reflect.Array:
		if source == nil && target.Elem().Kind() == reflect.Uint8 {
			return value.Interface(), true
		}
		items, ok := source.([]any)
		if target.Elem().Kind() == reflect.Uint8 {
			var bytes []byte
			switch source := source.(type) {
			case []byte:
				bytes = source
			case string:
				bytes = []byte(source)
			default:
				return nil, false
			}
			if len(bytes) > target.Len() {
				return nil, false
			}
			for index, item := range bytes {
				value.Index(index).SetUint(uint64(item))
			}
			return value.Interface(), true
		}
		if !ok && source != nil || len(items) > target.Len() {
			return nil, false
		}
		keys := make([]any, target.Len())
		for index := range keys {
			var item any
			if index < len(items) {
				item = items[index]
			}
			var supported bool
			keys[index], supported = projectedKey(item, target.Elem(), loose)
			if !supported {
				return nil, false
			}
		}
		return keys, true
	default:
		return nil, false
	}
	return value.Interface(), true
}
