# Quickstart

## Install

The project is currently unreleased. Until the first tag exists, pin a commit:

```sh
go get github.com/faustbrian/go-wire@<commit>
```

The module supports Go 1.25 and newer.

## Decode JSON

```go
var response struct {
	Status string `json:"status"`
}

err := jsonwire.DecodeReader(responseBody, &response, jsonwire.DecodeOptions{
	MaxBytes:              2 << 20,
	DisallowUnknownFields: true,
})
```

`MaxBytes: 0` selects the safe 1 MiB default. Decoding accepts exactly one JSON
value. Syntax errors classify as `wire.ErrParse`; unknown fields, type
mismatches, invalid targets, and size failures classify as
`wire.ErrValidation`.

## Encode JSON

```go
payload, err := jsonwire.Encode(response, jsonwire.EncodeOptions{})

err = jsonwire.EncodeWriter(destination, response, jsonwire.EncodeOptions{
	Indent: "  ",
})
```

`Encode` returns a complete JSON document as bytes. `EncodeWriter` writes the
same deterministic document without adding a trailing newline.

## Normalize JSON

```go
canonical, err := jsonwire.Normalize(vendorPayload, jsonwire.NormalizeOptions{})
```

Normalization strips leading/trailing whitespace and a leading UTF-8 BOM,
validates one JSON value, orders object keys through `encoding/json`, compacts
the document, and preserves number lexemes such as `1.20`. It does not reject
duplicate object keys.

## Decode namespace-aware XML

```go
var shipment struct {
	XMLName xml.Name `xml:"urn:vendor Shipment"`
	ID      int      `xml:"urn:vendor ID"`
}

err := xmlwire.Decode(payload, &shipment, xmlwire.DecodeOptions{
	ExpectedRoot: xml.Name{Space: "urn:vendor", Local: "Shipment"},
})
```

Strict XML is the default. Set `AllowNonStrict` only for a known peer whose
malformation is intentionally accepted. Prefixes are not identities;
`ExpectedRoot` compares the resolved namespace URL and local name.

## Encode XML

```go
payload, err := xmlwire.Encode(shipment, xmlwire.EncodeOptions{
	IncludeHeader: true,
})

err = xmlwire.EncodeWriter(destination, shipment, xmlwire.EncodeOptions{})
```

Both APIs serialize typed values through `encoding/xml`. The writer receives a
complete document, including the XML declaration when requested.

## Encode SOAP

```go
payload, err := soap.Encode(
	soap.Version12,
	header,
	request,
	soap.EncodeOptions{},
)

err = soap.EncodeWriter(
	destination,
	soap.Version12,
	header,
	request,
	soap.EncodeOptions{},
)
```

Pass `nil` for an omitted Header or an empty Body. For already encoded XML
fragments, use `Marshal` or `MarshalWriter`. `MarshalFault` and
`MarshalFaultWriter` provide the corresponding SOAP fault output paths.

## Parse SOAP

```go
envelope, err := soap.ParseReader(responseBody, soap.ParseOptions{})
if errors.Is(err, wire.ErrSOAPFault) {
	var faultError *soap.FaultError
	if errors.As(err, &faultError) {
		return fmt.Errorf("carrier %s: %s", faultError.Fault.Code, faultError.Fault.Reason)
	}
}
if err != nil {
	return err
}

var response RateResponse
if err := envelope.DecodeBody(&response); err != nil {
	return err
}
```

`Parse` and `ParseReader` return a non-nil envelope together with a
`*soap.FaultError` for a valid SOAP fault. Do not discard the envelope before
checking the error classification.

## Classify errors

```go
switch {
case errors.Is(err, wire.ErrSOAPFault):
	// The peer returned a valid fault response.
case errors.Is(err, wire.ErrEnvelope):
	// SOAP transport-envelope structure was invalid.
case errors.Is(err, wire.ErrParse):
	// Input was not valid JSON/XML/SOAP syntax.
case errors.Is(err, wire.ErrValidation):
	// Input syntax was valid but did not meet the requested shape or policy.
case errors.Is(err, wire.ErrUnsupportedFormat):
	// Explicit detection did not recognize JSON or XML.
case errors.Is(err, wire.ErrWrite):
	// Serialization succeeded but the destination rejected the output.
}
```

Use `errors.As(err, &wireError)` to inspect `Kind`, `Format`, `Op`, and the
underlying cause.
