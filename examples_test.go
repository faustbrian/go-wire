package wire_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/faustbrian/go-wire"
	"github.com/faustbrian/go-wire/jsonwire"
	"github.com/faustbrian/go-wire/soap"
	"github.com/faustbrian/go-wire/xmlwire"
)

func ExampleDetectFormat() {
	format, err := wire.DetectFormat([]byte(`{"status":"ok"}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(format)
	// Output: json
}

func Example_json() {
	var response struct {
		Status string `json:"status"`
	}
	err := jsonwire.DecodeReader(
		strings.NewReader(`{"status":"ok"}`),
		&response,
		jsonwire.DecodeOptions{DisallowUnknownFields: true},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response.Status)
	// Output: ok
}

func Example_xml() {
	var shipment struct {
		XMLName xml.Name `xml:"urn:vendor Shipment"`
		ID      int      `xml:"urn:vendor ID"`
	}
	payload := []byte(`<v:Shipment xmlns:v="urn:vendor"><v:ID>42</v:ID></v:Shipment>`)
	err := xmlwire.Decode(payload, &shipment, xmlwire.DecodeOptions{
		ExpectedRoot: xml.Name{Space: "urn:vendor", Local: "Shipment"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(shipment.ID)
	// Output: 42
}

func Example_soapFault() {
	payload, err := soap.MarshalFault(soap.Fault{
		Version: soap.Version11,
		Code:    "soap:Server",
		Reason:  "Carrier unavailable",
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = soap.Parse(payload, soap.ParseOptions{})
	if errors.Is(err, wire.ErrSOAPFault) {
		fmt.Println("fault")
	}
	// Output: fault
}

func Example_jsonWriter() {
	var output bytes.Buffer
	err := jsonwire.EncodeWriter(&output, map[string]int{"z": 2, "a": 1}, jsonwire.EncodeOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output.String())
	// Output: {"a":1,"z":2}
}

func Example_xmlWriter() {
	var output bytes.Buffer
	value := struct {
		XMLName xml.Name `xml:"Status"`
		Value   string   `xml:",chardata"`
	}{Value: "ok"}
	if err := xmlwire.EncodeWriter(&output, value, xmlwire.EncodeOptions{}); err != nil {
		log.Fatal(err)
	}
	fmt.Println(output.String())
	// Output: <Status>ok</Status>
}

func Example_soapWriter() {
	var output bytes.Buffer
	request := struct {
		XMLName xml.Name `xml:"Ping"`
	}{}
	if err := soap.EncodeWriter(&output, soap.Version11, nil, request, soap.EncodeOptions{}); err != nil {
		log.Fatal(err)
	}
	fmt.Println(output.String())
	// Output: <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Ping></Ping></soap:Body></soap:Envelope>
}
