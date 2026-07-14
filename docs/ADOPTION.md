# Adoption guide

## 1. Identify the boundary

Adopt `go-wire` where bytes enter or leave a service: an HTTP handler, carrier
client, webhook consumer, file import, or RPC adapter. Keep domain mapping
outside the package.

```text
transport bytes → go-wire parse/validate → boundary DTO → domain mapping
```

Do not pass `soap.Envelope`, raw maps, or vendor XML structs deep into business
logic unless raw wire data is itself a domain requirement.

## 2. Choose explicit format APIs

Use `jsonwire`, `xmlwire`, or `soap` when the integration contract names a
format. Use `wire.DetectFormat` only for genuinely polymorphic inputs such as a
diagnostic import tool. Detection does not validate the payload and reports
SOAP as XML.

## 3. Set payload limits from evidence

The 1 MiB default is a safe starting point, not a universal recommendation.
Measure representative production payloads and set the smallest limit that
allows documented headroom.

```go
const carrierResponseLimit = 2 << 20

err := jsonwire.DecodeReader(body, &response, jsonwire.DecodeOptions{
	MaxBytes: carrierResponseLimit,
})
```

Treat a size change as operational policy. Record it beside the integration,
not in a global magic constant shared by unrelated peers.

## 4. Decide strictness per peer

- Enable JSON unknown-field rejection when the upstream contract is closed and
  an added field should trigger review. Leave it disabled for additive APIs.
- Keep XML strict unless a fixture proves a specific vendor defect that Go can
  recover safely.
- Validate XML roots when routing or target selection depends on the namespace.
- Never infer SOAP 1.1 versus 1.2 from HTTP headers alone; parse the envelope
  namespace.

## 5. Map errors at the boundary

Log or report the shared classification, operation, and safe peer context.
Avoid logging entire payloads by default because they can contain credentials
or personal data.

```go
var wireError *wire.Error
if errors.As(err, &wireError) {
	logger.Error("carrier payload rejected",
		"format", wireError.Format,
		"operation", wireError.Op,
		"kind", wireError.Kind,
	)
}
```

Handle `wire.ErrSOAPFault` as a valid remote response, not corrupt XML. Decide
which fault codes are retryable in application code; the package does not know
the peer's retry semantics.

## 6. Preserve fixtures

For each adopted integration, keep redacted fixtures for:

- the smallest valid response;
- a representative large response;
- every accepted vendor quirk;
- malformed syntax;
- a valid shape with invalid content;
- SOAP faults for both retryable and terminal cases, where applicable.

Test the boundary DTO and error classification. A regression fixture should
explain which real interoperability behavior it protects.

## 7. Roll out safely

1. Parse shadow copies in tests or a non-production replay harness.
2. Compare domain DTOs and emitted bytes with the existing implementation.
3. Investigate every divergence; do not normalize it away without documenting
   the behavior.
4. Deploy behind the application's normal release controls.
5. Monitor classification counts and payload-limit failures without recording
   sensitive bodies.

## 8. Upgrade deliberately

Pin tagged versions after v1. Read `CHANGELOG.md` before every upgrade. Defaults,
normalization, SOAP fault mapping, emitted bytes, and error classification are
compatibility-sensitive even if function signatures do not change.
