# Supported formats and behavior matrix

## Format support

| Capability | JSON | XML | SOAP 1.1 | SOAP 1.2 |
| --- | --- | --- | --- | --- |
| Bounded byte decode | Yes | Yes | Yes | Yes |
| Bounded reader decode | Yes | Yes | Yes | Yes |
| Deterministic encode | Yes | Yes for typed values | Yes | Yes |
| Typed encode to `io.Writer` | Yes | Yes | Yes | Yes |
| Strict syntax default | Yes | Yes | Yes | Yes |
| Raw payload access | Caller retains input | Caller retains input | Envelope/header/body | Envelope/header/body |
| Namespace-aware validation | N/A | Root name | Envelope and body | Envelope, body, fault |
| Typed faults | N/A | N/A | Yes | Yes, including localized reasons/subcodes |
| Explicit normalization | BOM/whitespace/key order | No generic normalizer | Raw fragments preserved | Raw fragments preserved |
| Fuzz target | Decode | Decode | Parse | Parse |
| Parse/encode benchmarks | Yes | Yes | Yes | Yes |

## Error behavior

| Situation | Classification |
| --- | --- |
| Broken JSON/XML syntax or failed body read | `wire.ErrParse` |
| Unknown strict JSON field or target type mismatch | `wire.ErrValidation` |
| Invalid target, option, or payload size | `wire.ErrValidation` |
| Unknown explicit detected format | `wire.ErrUnsupportedFormat` |
| Wrong SOAP namespace, ordering, cardinality, or fault shape | `wire.ErrEnvelope` |
| Valid SOAP fault response | `wire.ErrSOAPFault` and `*soap.FaultError` |
| Output destination rejects encoded bytes | `wire.ErrWrite` |

## Charset support

| Label family | XML | SOAP | Notes |
| --- | --- | --- | --- |
| UTF-8 / UTF8 | Built in | Built in | Invalid UTF-8 is rejected |
| US-ASCII / ASCII | Built in | Built in | Bytes above `0x7f` are rejected |
| ISO-8859-1 / Latin-1 | Built in | Built in | Every byte maps directly to Unicode |
| Windows-1252 / CP1252 | Built in | Built in | Undefined bytes are rejected |
| Other declared charset | Inject reader | Inject reader | No guessing or fallback |

## Intentional limitations

### JSON

- Duplicate object names follow `encoding/json` behavior; later values can
  replace or merge into earlier values depending on the target.
- No JSON Schema, JSON-RPC, JSON:API, JSON Patch, streaming token facade, or
  canonical JSON standard is implemented.
- `Normalize` is presentation normalization, not cryptographic
  canonicalization.

### XML

- No DTD validation, XSD validation, XPath, XSLT, canonical XML, digital
  signatures, entity fetching, or arbitrary charset registry.
- `AllowNonStrict` inherits `encoding/xml` recovery behavior and can change the
  interpreted tree. It is never enabled implicitly.
- Encoding may choose namespace declarations different from a source document.

### SOAP

- No WSDL generation, service proxy, SOAPAction policy, MTOM, attachments,
  WS-Addressing, WS-Security, XML signatures, retries, or HTTP client behavior.
- A fault must be the only direct Body child and must contain a code and reason.
- `DecodeBody` requires exactly one direct Body child.
- Raw fragment marshaling checks XML well-formedness but does not apply an
  application schema.

### All formats

- Default limits are 1 MiB and are part of the compatibility contract.
- Transport status, headers, timeouts, cancellation, retries, and logging are
  owned by the calling application.
- YAML, TOML, MessagePack, queues, persistence, and business mapping are out of
  scope for v1.
