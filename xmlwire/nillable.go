package xmlwire

import (
	"encoding"
	"encoding/xml"
	"reflect"
	"strings"
)

// SchemaInstanceNamespace is the XML Schema instance namespace that owns the
// xsi:nil attribute.
const SchemaInstanceNamespace = "http://www.w3.org/2001/XMLSchema-instance"

// Nillable keeps the three states an optional, nillable XML Schema element can
// have: absent (Present false), explicitly nil (Present and Null true) and a
// value. Decoding honors xsi:nil only in SchemaInstanceNamespace; a nil element
// must have no child elements and no text other than XML whitespace. A
// non-nil element decodes through encoding/xml, except that empty content for
// a numeric or boolean T is rejected instead of becoming zero or false, so a
// blank value cannot pass as a real one; string types keep empty content.
// Encoding writes nothing when Present is false, writes an empty element whose
// nil attribute is in SchemaInstanceNamespace (encoding/xml chooses the
// prefix) when Null is true, and otherwise encodes Value with the element's
// name.
type Nillable[T any] struct {
	// Present reports that the element appeared.
	Present bool
	// Null reports an explicit xsi:nil="true" or "1".
	Null bool
	// Value is the decoded content when the element is present and not nil.
	Value T
}

// UnmarshalXML decodes one element into the nillable value.
func (value *Nillable[T]) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var zero T
	*value = Nillable[T]{Present: true, Value: zero}
	null, err := nilMarker(start)
	if err != nil {
		return err
	}
	if !null {
		if !rejectsEmptyContent(reflect.TypeFor[T]()) {
			return decoder.DecodeElement(&value.Value, &start)
		}
		var text string
		if err := decoder.DecodeElement(&text, &start); err != nil {
			return err
		}
		if trimXMLSpace(text) == "" {
			return xml.UnmarshalError("element " + start.Name.Local + " has no value")
		}
		// Re-decode the checked text so numeric and boolean parsing stays
		// encoding/xml's own.
		var escaped strings.Builder
		_ = xml.EscapeText(&escaped, []byte(text))
		return xml.Unmarshal([]byte("<v>"+escaped.String()+"</v>"), &value.Value)
	}
	value.Null = true
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.EndElement:
			return nil
		case xml.CharData:
			if trimXMLSpace(string(token)) != "" {
				return xml.UnmarshalError("nil element " + start.Name.Local + " has text content")
			}
		case xml.StartElement:
			return xml.UnmarshalError("nil element " + start.Name.Local + " has child elements")
		}
	}
}

// MarshalXML encodes the nillable value, or nothing when it is absent.
func (value Nillable[T]) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if !value.Present {
		return nil
	}
	if !value.Null {
		return encoder.EncodeElement(value.Value, start)
	}
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: SchemaInstanceNamespace, Local: "nil"}, Value: "true"})
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}
	return encoder.EncodeToken(start.End())
}

// nilMarker reads xsi:nil using the XML Schema boolean lexical space after
// whitespace collapse; any other value is a validation failure.
func nilMarker(start xml.StartElement) (bool, error) {
	for _, attribute := range start.Attr {
		if attribute.Name.Space != SchemaInstanceNamespace || attribute.Name.Local != "nil" {
			continue
		}
		switch trimXMLSpace(attribute.Value) {
		case "true", "1":
			return true, nil
		case "false", "0":
			return false, nil
		default:
			return false, xml.UnmarshalError("xsi:nil on " + start.Name.Local + " is not an XML Schema boolean")
		}
	}
	return false, nil
}

// trimXMLSpace removes XML whitespace (space, tab, carriage return and line
// feed) only, as XML Schema whitespace collapse does; other Unicode spaces are
// content.
func trimXMLSpace(text string) string {
	return strings.Trim(text, " \t\r\n")
}

var (
	xmlUnmarshalerType  = reflect.TypeFor[xml.Unmarshaler]()
	textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
)

// rejectsEmptyContent reports a numeric or boolean value, possibly behind
// pointers, that encoding/xml would decode from empty content as zero or
// false. A type with its own XML or text unmarshaling decides for itself.
func rejectsEmptyContent(kind reflect.Type) bool {
	for {
		if customUnmarshaling(kind) {
			return false
		}
		if kind.Kind() != reflect.Pointer {
			return scalarKind(kind)
		}
		kind = kind.Elem()
	}
}

func customUnmarshaling(kind reflect.Type) bool {
	pointer := reflect.PointerTo(kind)
	return kind.Implements(xmlUnmarshalerType) || kind.Implements(textUnmarshalerType) ||
		pointer.Implements(xmlUnmarshalerType) || pointer.Implements(textUnmarshalerType)
}

func scalarKind(kind reflect.Type) bool {
	switch kind.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
