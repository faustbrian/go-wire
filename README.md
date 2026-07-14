# go-wire

`go-wire` is a transport-neutral Go package for explicit, auditable handling
of JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON interoperability
boundaries, with bounded read and symmetric write APIs.

The project is under active development toward its first release. APIs in the
unreleased series are not yet covered by a stable-version compatibility
promise.

go-wire requires Go 1.25.8 or newer.

## Support matrix

| Format | Current support | Intentional limits |
| --- | --- | --- |
| JSON | Bounded UTF-8 decode, deterministic encode, BOM/whitespace normalization | Duplicate object keys follow `encoding/json`; no schema validation |
| XML | Byte/depth-bounded strict decode, deterministic encode, namespace/root validation, four common charsets | No DTD/schema validation or arbitrary charset registry |
| SOAP | Byte/depth-bounded SOAP 1.1/1.2 parse and emit, raw sections, body decode, typed faults | No WSDL, WS-Security, WS-Addressing, or HTTP client policy |
| YAML | Bounded YAML 1.2 read/write, duplicate and multi-document policy, alias/depth limits | YAML v4 compatibility rules apply; no schema or automatic detection |
| TOML | Complete-document read/write, strict unknown fields, native datetime and numeric validation | No streaming multi-document form or automatic detection |
| MessagePack | One-object read/write, sorted maps, timestamps, exact numeric widths, unique keys, bounded nesting/collections | Untyped maps require string keys; unknown extensions are rejected |
| CBOR | Canonical/Core/CTAP2 deterministic output, bounded collections, explicit tags and indefinite lengths | Tags and indefinite lengths are rejected by default |
| BSON | Complete documents, recursive duplicate checks, ObjectID/datetime/raw documents | No top-level scalars; `M` map output order is not guaranteed |

Generic RPC behavior, automatic binary detection, HTTP client policy,
persistence, and application-specific mappings are not supported.

## Design

The root `wire` package owns only cross-format primitives. Format-specific
behavior lives in explicit packages so call sites do not hide wire semantics
behind a generic codec abstraction.

```go
package main

import (
	"fmt"
	"log"

	"github.com/faustbrian/go-wire"
)

func main() {
	payload := []byte(`{"status":"ok"}`)

	format, err := wire.DetectFormat(payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(format)
}
```

Every format can return encoded bytes with `Encode` or write a complete value
with `EncodeWriter`. `EncodeOptions.MaxBytes` bounds output and zero selects the
safe 1 MiB default. SOAP supports the same typed flow for both the optional
Header and Body:

```go
err := soap.EncodeWriter(
	requestBody,
	soap.Version12,
	header,
	request,
	soap.EncodeOptions{MaxBytes: 256 << 10},
)
```

Detection is optional. It reports SOAP bytes as XML because trustworthy SOAP
identification requires parsing and validating the envelope namespace.
YAML and TOML are intentionally not guessed from arbitrary text, and
MessagePack, CBOR, and BSON are never guessed from arbitrary bytes.

JSON callers can enforce a bounded body and reject fields that their target
does not declare:

```go
package main

import (
	"log"
	"strings"

	"github.com/faustbrian/go-wire/jsonwire"
)

func main() {
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
}
```

XML root validation compares fully resolved namespace URLs, not source
prefixes:

```go
var shipment struct {
	XMLName xml.Name `xml:"urn:vendor Shipment"`
	ID      int      `xml:"urn:vendor ID"`
}

err := xmlwire.Decode(payload, &shipment, xmlwire.DecodeOptions{
	ExpectedRoot: xml.Name{Space: "urn:vendor", Local: "Shipment"},
})
```

Strict parsing is the default. `AllowNonStrict` is an explicit interoperability
escape hatch and can repair mismatched or missing closing tags according to
`encoding/xml` behavior. The built-in charset reader supports UTF-8, US-ASCII,
ISO-8859-1, and Windows-1252 only; provide `DecodeOptions.CharsetReader` for an
additional known vendor encoding.

SOAP parsing validates the envelope namespace and structure while retaining
the exact bytes needed for diagnostics:

```go
envelope, err := soap.ParseReader(response.Body, soap.ParseOptions{})
if errors.Is(err, wire.ErrSOAPFault) {
	var faultError *soap.FaultError
	if errors.As(err, &faultError) {
		log.Printf("carrier fault %s: %s", faultError.Fault.Code, faultError.Fault.Reason)
	}
}
if err != nil {
	return err
}

var rate RateResponse
if err := envelope.DecodeBody(&rate); err != nil {
	return err
}
```

`Encode` and `EncodeWriter` turn typed header/body values into complete SOAP
envelopes. `Marshal` and `MarshalWriter` wrap validated raw fragments without
rewriting them. Their `WithOptions` variants configure the output bound.
Faults have matching byte and writer APIs. The package does not
infer a version from payload content and does not implement transport, WSDL,
WS-Security, or application mapping behavior.

## Documentation

- [Quickstart](docs/QUICKSTART.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Public API reference](docs/API.md)
- [Audit evidence and traceability](docs/EVIDENCE.md)
- [Supported formats and limitations](docs/FORMATS.md)
- [Codec dependencies and security posture](docs/DEPENDENCIES.md)
- [Hardening audit and conformance matrix](docs/HARDENING.md)
- [Performance and allocation policy](docs/PERFORMANCE.md)
- [Adoption guide](docs/ADOPTION.md)
- [End-to-end examples](docs/EXAMPLES.md)
- [Scenario cookbook](docs/COOKBOOK.md)
- [FAQ](docs/FAQ.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [Migration notes](docs/MIGRATION.md)
- [Versioning and release guide](docs/VERSIONING.md)
- [Roadmap](ROADMAP.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Project goal and v1 acceptance criteria](GOAL.md)
- [Changelog](CHANGELOG.md)
- [Portable AI documentation index](llms.txt)
- [Complete AI documentation bundle](llms-full.txt)

## Development

```sh
go test ./...
```

See [GOAL.md](GOAL.md) for the complete v1 scope and acceptance criteria.
