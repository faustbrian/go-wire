# Troubleshooting

## Start with classification

```go
var wireError *wire.Error
if errors.As(err, &wireError) {
	fmt.Printf("format=%s kind=%s operation=%s cause=%v\n",
		wireError.Format, wireError.Kind, wireError.Op, wireError.Err)
}
```

Use the cause for engineering diagnostics, but do not expose vendor payloads or
internal parse details directly to untrusted clients.

## `payload exceeds size limit`

The zero-value option uses 1 MiB. Confirm the peer's documented maximum and
measure redacted fixtures before raising the limit. A sudden size increase can
also indicate an error page, batch expansion, or upstream regression.

## JSON reports `multiple JSON values`

The payload contains another non-whitespace value after the first. Common causes
are newline-delimited JSON, concatenated responses, or a proxy error appended
to a valid document. NDJSON is not supported by `Decode`.

## JSON unknown fields fail

`DisallowUnknownFields` was enabled. Decide whether the upstream contract is
closed. If fields are additive, disable the option; do not repeatedly add dummy
fields solely to silence validation.

## JSON numbers look different after application decoding

`Normalize` preserves source number lexemes, but decoding into `float64` does
not. Use `json.Number`, a decimal type owned by the application, or a string
field when the peer's contract requires exact decimal representation.

## XML says the root namespace is wrong

Inspect `xmlwire.Root`. Compare `xml.Name.Space` to the namespace URL, not the
source prefix. An empty `Space` usually means the peer omitted `xmlns` or placed
the element outside the expected default namespace.

## XML reports an unsupported charset

Check the declaration and raw bytes. Correct a falsely declared upstream
encoding at the source where possible. Otherwise inject a narrowly scoped
`CharsetReader`; never guess based on byte appearance.

## Non-strict XML still fails

`AllowNonStrict` is the recovery behavior from `encoding/xml`, not a general
repair engine. Capture a redacted fixture and determine whether the document
can be repaired unambiguously. Reject it if multiple interpretations exist.

## SOAP is classified as an envelope failure

Check all of the following:

- root local name is `Envelope`;
- namespace is exactly the SOAP 1.1 or SOAP 1.2 namespace;
- at most one Header appears and it precedes Body;
- exactly one Body appears;
- envelope children use the envelope namespace;
- a Fault is the only direct Body child and has a code and reason.

Malformed XML is `wire.ErrParse`; well-formed XML that violates these rules is
`wire.ErrEnvelope`.

## `DecodeBody` says the body has zero or multiple children

The helper intentionally targets request/response SOAP shapes with one body
entry. Use `BodyXML` and a peer-specific XML model if the peer legitimately
uses multiple body entries, and document that divergence.

## A valid SOAP fault appears as an error

This is expected. Check `errors.Is(err, wire.ErrSOAPFault)` before treating the
response as broken XML. Use `errors.As` to obtain `*soap.FaultError`, or read
`envelope.Fault` from the simultaneously returned envelope.

## SOAP fragment marshaling fails

Header and body arguments must be XML fragments containing elements or
whitespace, not plain text or an unclosed element. Prefixes used inside a
fragment must be declared in that fragment unless they use the package's
`soap` envelope prefix.

## Coverage falls below 100%

Run:

```sh
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Add behavior-focused tests for the missing branch. Do not add assertions that
only execute a line without proving its contract.
