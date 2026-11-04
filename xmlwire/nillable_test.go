package xmlwire_test

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
	"testing"

	"github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/xmlwire"
)

type nillableRecord struct {
	XMLName  xml.Name                  `xml:"record"`
	Distance xmlwire.Nillable[float64] `xml:"Distance"`
	Terminal xmlwire.Nillable[string]  `xml:"Terminal"`
}

const xsiNamespace = `xmlns:i="http://www.w3.org/2001/XMLSchema-instance"`

func TestNillableDistinguishesAbsentNilEmptyAndValue(t *testing.T) {
	t.Parallel()

	var absent nillableRecord
	if err := xmlwire.Decode([]byte(`<record/>`), &absent, xmlwire.DecodeOptions{}); err != nil {
		t.Fatalf("Decode(absent) error = %v", err)
	}
	if absent.Distance.Present || absent.Distance.Null || absent.Terminal.Present {
		t.Fatalf("absent elements = %#v", absent)
	}

	for _, marker := range []string{"true", "1", " true "} {
		var null nillableRecord
		payload := `<record ` + xsiNamespace + `><Distance i:nil="` + marker + `"/><Terminal i:nil="` + marker + `"></Terminal></record>`
		if err := xmlwire.Decode([]byte(payload), &null, xmlwire.DecodeOptions{}); err != nil {
			t.Fatalf("Decode(nil=%q) error = %v", marker, err)
		}
		if !null.Distance.Present || !null.Distance.Null || !null.Terminal.Null || null.Terminal.Value != "" {
			t.Fatalf("nil elements (%q) = %#v", marker, null)
		}
	}

	var value nillableRecord
	payload := `<record ` + xsiNamespace + `><Distance i:nil="false">1.25</Distance><Terminal></Terminal></record>`
	if err := xmlwire.Decode([]byte(payload), &value, xmlwire.DecodeOptions{}); err != nil {
		t.Fatalf("Decode(value) error = %v", err)
	}
	if !value.Distance.Present || value.Distance.Null || value.Distance.Value != 1.25 {
		t.Fatalf("Distance = %#v", value.Distance)
	}
	if !value.Terminal.Present || value.Terminal.Null || value.Terminal.Value != "" {
		t.Fatalf("empty Terminal = %#v", value.Terminal)
	}
}

func TestNillableIgnoresNilOutsideTheSchemaInstanceNamespace(t *testing.T) {
	t.Parallel()

	var record nillableRecord
	payload := `<record xmlns:x="urn:other"><Terminal x:nil="true">open</Terminal><Distance nil="true">2</Distance></record>`
	if err := xmlwire.Decode([]byte(payload), &record, xmlwire.DecodeOptions{}); err != nil {
		t.Fatalf("Decode error = %v", err)
	}
	if record.Terminal.Null || record.Terminal.Value != "open" || record.Distance.Null || record.Distance.Value != 2 {
		t.Fatalf("record = %#v", record)
	}
}

func TestNillableRejectsContentInANilElementAndBadMarkers(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		`<Terminal i:nil="true">text</Terminal>`,
		`<Terminal i:nil="true"><child/></Terminal>`,
		`<Terminal i:nil="yes"/>`,
	} {
		var record nillableRecord
		err := xmlwire.Decode([]byte(`<record `+xsiNamespace+`>`+body+`</record>`), &record, xmlwire.DecodeOptions{})
		assertKind(t, err, wire.ErrValidation)
	}
	var record nillableRecord
	if err := xmlwire.Decode([]byte(`<record `+xsiNamespace+`><Terminal i:nil="true">  </Terminal></record>`), &record, xmlwire.DecodeOptions{}); err != nil || !record.Terminal.Null {
		t.Fatalf("whitespace-only nil element = %#v, %v", record.Terminal, err)
	}
}

func TestNillableEncodesValueNilAndAbsence(t *testing.T) {
	t.Parallel()

	encoded, err := xmlwire.Encode(nillableRecord{
		Distance: xmlwire.Nillable[float64]{Present: true, Value: 3.5},
		Terminal: xmlwire.Nillable[string]{Present: true, Null: true},
	}, xmlwire.EncodeOptions{})
	if err != nil {
		t.Fatalf("Encode error = %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, `<Distance>3.5</Distance>`) || !strings.Contains(text, `nil="true"`) || !strings.Contains(text, `http://www.w3.org/2001/XMLSchema-instance`) {
		t.Fatalf("Encode = %s", text)
	}

	var decoded nillableRecord
	if err := xmlwire.Decode(encoded, &decoded, xmlwire.DecodeOptions{}); err != nil {
		t.Fatalf("Decode(Encode) error = %v", err)
	}
	if decoded.Distance.Value != 3.5 || !decoded.Terminal.Null {
		t.Fatalf("round trip = %#v", decoded)
	}

	omitted, err := xmlwire.Encode(nillableRecord{}, xmlwire.EncodeOptions{})
	if err != nil || string(omitted) != `<record></record>` {
		t.Fatalf("Encode(absent) = %s, %v", omitted, err)
	}
}

func TestNillableRefusesToEncodeAPresentNilPointerAsAbsent(t *testing.T) {
	t.Parallel()

	type pointerRecord struct {
		XMLName  xml.Name                   `xml:"r"`
		Distance xmlwire.Nillable[*float64] `xml:"d"`
	}
	encoded, err := xmlwire.Encode(pointerRecord{Distance: xmlwire.Nillable[*float64]{Present: true}}, xmlwire.EncodeOptions{})
	assertKind(t, err, wire.ErrValidation)
	if encoded != nil {
		t.Fatalf("Encode(present nil pointer) wrote %s", encoded)
	}
	distance := 2.5
	encoded, err = xmlwire.Encode(pointerRecord{Distance: xmlwire.Nillable[*float64]{Present: true, Value: &distance}}, xmlwire.EncodeOptions{})
	if err != nil || string(encoded) != `<r><d>2.5</d></r>` {
		t.Fatalf("Encode(present pointer) = %s, %v", encoded, err)
	}
}

func TestNillableRejectsEmptyNumericAndBooleanContent(t *testing.T) {
	t.Parallel()

	for _, body := range []string{`<Distance/>`, `<Distance>  </Distance>`} {
		var record nillableRecord
		assertKind(t, xmlwire.Decode([]byte(`<record>`+body+`</record>`), &record, xmlwire.DecodeOptions{}), wire.ErrValidation)
	}
	var flag struct {
		XMLName xml.Name               `xml:"r"`
		Ready   xmlwire.Nillable[bool] `xml:"ready"`
	}
	assertKind(t, xmlwire.Decode([]byte(`<r><ready></ready></r>`), &flag, xmlwire.DecodeOptions{}), wire.ErrValidation)
	if err := xmlwire.Decode([]byte(`<r><ready> true </ready></r>`), &flag, xmlwire.DecodeOptions{}); err != nil || !flag.Ready.Value {
		t.Fatalf("Decode(bool) = %#v, %v", flag.Ready, err)
	}
	var record nillableRecord
	if err := xmlwire.Decode([]byte(`<record><Distance>-0.5</Distance></record>`), &record, xmlwire.DecodeOptions{}); err != nil || record.Distance.Value != -0.5 {
		t.Fatalf("Decode(number) = %#v, %v", record.Distance, err)
	}
	assertKind(t, xmlwire.Decode([]byte(`<record><Distance>x</Distance></record>`), &record, xmlwire.DecodeOptions{}), wire.ErrValidation)
}

func TestNillableTreatsOnlyXMLWhitespaceAsSpace(t *testing.T) {
	t.Parallel()

	var record nillableRecord
	payload := "<record " + xsiNamespace + "><Terminal i:nil=\"\u00a0true\"/></record>"
	assertKind(t, xmlwire.Decode([]byte(payload), &record, xmlwire.DecodeOptions{}), wire.ErrValidation)
	payload = "<record " + xsiNamespace + "><Terminal i:nil=\"true\">\u00a0</Terminal></record>"
	assertKind(t, xmlwire.Decode([]byte(payload), &record, xmlwire.DecodeOptions{}), wire.ErrValidation)
}

// lenientCount owns its empty-content meaning through TextUnmarshaler.
type lenientCount int

func (count *lenientCount) UnmarshalText(text []byte) error {
	if len(bytes.TrimSpace(text)) == 0 {
		*count = -1
		return nil
	}
	value, err := strconv.Atoi(string(bytes.TrimSpace(text)))
	*count = lenientCount(value)
	return err
}

func TestNillableLeavesEmptyContentToCustomUnmarshalers(t *testing.T) {
	t.Parallel()

	var record struct {
		XMLName xml.Name                       `xml:"r"`
		Count   xmlwire.Nillable[lenientCount] `xml:"count"`
	}
	if err := xmlwire.Decode([]byte(`<r><count/></r>`), &record, xmlwire.DecodeOptions{}); err != nil || record.Count.Value != -1 {
		t.Fatalf("Decode(custom empty) = %#v, %v", record.Count, err)
	}
	if err := xmlwire.Decode([]byte(`<r><count>7</count></r>`), &record, xmlwire.DecodeOptions{}); err != nil || record.Count.Value != 7 {
		t.Fatalf("Decode(custom value) = %#v, %v", record.Count, err)
	}
}

func TestNillableRejectsEmptyContentBehindPointers(t *testing.T) {
	t.Parallel()

	var record struct {
		XMLName  xml.Name                   `xml:"r"`
		Distance xmlwire.Nillable[*float64] `xml:"d"`
	}
	assertKind(t, xmlwire.Decode([]byte(`<r><d/></r>`), &record, xmlwire.DecodeOptions{}), wire.ErrValidation)
	if err := xmlwire.Decode([]byte(`<r><d>2.5</d></r>`), &record, xmlwire.DecodeOptions{}); err != nil || record.Distance.Value == nil || *record.Distance.Value != 2.5 {
		t.Fatalf("Decode(pointer) = %#v, %v", record.Distance, err)
	}
}
