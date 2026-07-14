# Scenario cookbook

## Reject additive JSON fields

Use this only when an upstream contract is intentionally closed:

```go
err := jsonwire.Decode(payload, &target, jsonwire.DecodeOptions{
	DisallowUnknownFields: true,
})
```

An unknown field is `wire.ErrValidation`, not a syntax error.

## Preserve JSON decimal spelling during normalization

```go
normalized, err := jsonwire.Normalize([]byte(`{"amount":1.20}`), jsonwire.NormalizeOptions{})
// normalized is {"amount":1.20}
```

Normalization retains `json.Number` lexemes. Decoding into `float64` elsewhere
can still change representation or precision.

## Disable JSON HTML escaping

```go
payload, err := jsonwire.Encode(value, jsonwire.EncodeOptions{
	DisableHTMLEscaping: true,
})
```

Use this only when literal `<`, `>`, and `&` are required by a peer. It is not a
general HTML-safety recommendation.

## Accept a known non-strict XML document

```go
err := xmlwire.Decode(payload, &target, xmlwire.DecodeOptions{
	AllowNonStrict: true,
})
```

Keep the malformed vendor payload as a regression fixture. Document the exact
repair that `encoding/xml` performs and review it when upgrading Go.

## Validate an XML namespace independent of prefix

```go
options := xmlwire.DecodeOptions{
	ExpectedRoot: xml.Name{Space: "urn:vendor:v2", Local: "Shipment"},
}
```

`v:Shipment` and `shipment:Shipment` are equivalent when both prefixes resolve
to `urn:vendor:v2`.

## Add a vendor charset

```go
options := xmlwire.DecodeOptions{
	CharsetReader: func(label string, input io.Reader) (io.Reader, error) {
		if strings.EqualFold(label, "vendor-ebcdic") {
			return newVendorDecoder(input), nil
		}
		return xmlwire.CharsetReader(label, input)
	},
}
```

Delegate built-in labels back to `xmlwire.CharsetReader`. Never silently treat
an unknown label as UTF-8.

## Inspect a SOAP fault without losing the envelope

```go
envelope, err := soap.Parse(payload, soap.ParseOptions{})
if errors.Is(err, wire.ErrSOAPFault) {
	fault := envelope.Fault
	// Use fault.Code, fault.Reason, fault.Detail, and envelope.RawXML().
}
```

The envelope is intentionally returned with the fault error.

## Emit localized SOAP 1.2 fault reasons

```go
payload, err := soap.MarshalFault(soap.Fault{
	Version: soap.Version12,
	Code:    "soap:Sender",
	Reasons: []soap.FaultReason{
		{Language: "en", Text: "Invalid shipment"},
		{Language: "fi", Text: "Virheellinen lähetys"},
	},
})
```

The first reason becomes the convenience `Fault.Reason` when parsed.

## Preserve a vendor body fragment exactly

```go
payload, err := soap.Marshal(soap.Version11, headerFragment, bodyFragment)
```

Fragments are checked for XML well-formedness and inserted byte-for-byte. Their
application schema and namespace meaning are not validated.

## Distinguish a size failure from malformed syntax

```go
if errors.Is(err, jsonwire.ErrPayloadTooLarge) {
	// JSON limit was exceeded.
}
if errors.Is(err, wire.ErrValidation) {
	// Includes size and shape policy failures.
}
```

XML exposes the corresponding `xmlwire.ErrPayloadTooLarge`. SOAP size failures
use the shared validation classification.
