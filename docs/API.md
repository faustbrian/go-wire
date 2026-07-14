# Public API reference

This document describes every exported v1-candidate symbol. Go signatures and
package comments remain authoritative; use `go doc` for locally installed
source.

## Package `wire`

### Formats

- `type Format string` identifies a supported format.
- `FormatJSON`, `FormatXML`, and `FormatSOAP` are the explicit identifiers.
- `DetectFormat([]byte) (Format, error)` detects JSON objects/arrays or XML by
  the first significant byte. It removes whitespace and one UTF-8 BOM. SOAP is
  reported as XML.

### Errors

- `type ErrorKind string` classifies failures.
- `ErrorKindParse`, `ErrorKindValidation`, `ErrorKindUnsupported`,
  `ErrorKindEnvelope`, and `ErrorKindFault` are stable classifications.
- `ErrParse`, `ErrValidation`, `ErrUnsupportedFormat`, `ErrEnvelope`, and
  `ErrSOAPFault` are sentinels for `errors.Is`.
- `type Error struct { Kind ErrorKind; Format Format; Op string; Err error }`
  carries structured context.
- `(*Error).Error() string` renders a stable contextual message.
- `(*Error).Is(error) bool` matches the sentinel for `Kind` or an underlying
  cause.
- `(*Error).Unwrap() error` returns the underlying cause.

## Package `jsonwire`

- `DefaultMaxBytes` is 1 MiB.
- `ErrPayloadTooLarge` identifies a configured limit violation and is wrapped
  by `wire.ErrValidation`.
- `DecodeOptions` contains `MaxBytes` and `DisallowUnknownFields`.
- `EncodeOptions` contains `Indent` and `DisableHTMLEscaping`.
- `NormalizeOptions` contains `MaxBytes`.
- `Decode([]byte, any, DecodeOptions) error` decodes exactly one value into a
  non-nil pointer.
- `DecodeReader(io.Reader, any, DecodeOptions) error` adds bounded stream
  reading.
- `Encode(any, EncodeOptions) ([]byte, error)` emits deterministic JSON.
- `Normalize([]byte, NormalizeOptions) ([]byte, error)` validates and emits
  compact key-ordered JSON while preserving number lexemes.

## Package `xmlwire`

- `DefaultMaxBytes` is 1 MiB.
- `ErrPayloadTooLarge` identifies a configured limit violation and is wrapped
  by `wire.ErrValidation`.
- `DecodeOptions` contains `MaxBytes`, `AllowNonStrict`, `ExpectedRoot`, and an
  injectable `CharsetReader`.
- `EncodeOptions` contains `Indent` and `IncludeHeader`.
- `Decode([]byte, any, DecodeOptions) error` validates and decodes one complete
  XML document.
- `DecodeReader(io.Reader, any, DecodeOptions) error` adds bounded stream
  reading.
- `Root([]byte, DecodeOptions) (xml.Name, error)` validates a complete document
  and returns the resolved root namespace and local name.
- `Encode(any, EncodeOptions) ([]byte, error)` serializes through
  `encoding/xml`.
- `CharsetReader(string, io.Reader) (io.Reader, error)` converts UTF-8,
  US-ASCII, ISO-8859-1, or Windows-1252 to UTF-8.

## Package `soap`

### Versions and options

- `type Version string`, `Version11`, and `Version12` identify the supported
  envelope namespaces.
- `DefaultMaxBytes` is 1 MiB.
- `ParseOptions` contains `MaxBytes` and an injectable `CharsetReader`.

### Envelopes

- `Parse([]byte, ParseOptions) (*Envelope, error)` validates an envelope.
- `ParseReader(io.Reader, ParseOptions) (*Envelope, error)` reads a bounded
  stream before validation.
- `Marshal(Version, header, body []byte) ([]byte, error)` wraps validated raw
  fragments in a deterministic envelope.
- `Envelope.Version` is the parsed SOAP version.
- `Envelope.Fault` is non-nil for a valid fault body.
- `Envelope.RawXML()`, `HeaderXML()`, and `BodyXML()` return defensive copies.
- `Envelope.DecodeBody(any) error` decodes exactly one body child with inherited
  namespace bindings.

### Faults

- `FaultReason` contains a SOAP 1.2 language and text pair.
- `Fault` contains `Version`, `Code`, `Subcodes`, `Reason`, `Reasons`, `Actor`,
  `Node`, `Role`, `Detail`, and `Raw`.
- `FaultError` carries a `Fault`, formats its code/reason, and unwraps to the
  shared SOAP fault classification.
- `(*FaultError).Error() string` renders the fault code and optional reason.
- `(*FaultError).Unwrap() error` exposes a `wire.Error` classified as
  `wire.ErrSOAPFault`.
- `MarshalFault(Fault) ([]byte, error)` validates and emits a complete SOAP
  fault envelope.

## Compatibility note

These symbols are release candidates until v1.0.0. After v1, exported names,
signatures, default limits, error classifications, wire output, and documented
normalization behavior are SemVer-governed.
