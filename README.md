# go-wire

`go-wire` is a transport-neutral Go package for explicit, auditable handling
of JSON, XML, and SOAP interoperability boundaries.

The project is under active development toward its first release. APIs in the
unreleased series are not yet covered by a stable-version compatibility
promise.

## Support matrix

| Format | Current support | Intentional limits |
| --- | --- | --- |
| JSON | Bounded decode, deterministic encode, BOM/whitespace normalization | Duplicate object keys follow `encoding/json`; no schema validation |
| XML | Bounded strict decode, deterministic encode, namespace/root validation, four common charsets | No DTD/schema validation or arbitrary charset registry |
| SOAP | SOAP 1.1/1.2 parse and emit, raw sections, body decode, typed faults | No WSDL, WS-Security, WS-Addressing, or HTTP client policy |

YAML, TOML, MessagePack, generic RPC behavior, HTTP client policy, persistence,
and application-specific mappings are not supported.

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

Detection is optional. It reports SOAP bytes as XML because trustworthy SOAP
identification requires parsing and validating the envelope namespace.

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

`Marshal` wraps validated raw header and body fragments without rewriting them.
`MarshalFault` emits version-correct SOAP 1.1 or SOAP 1.2 fault structure. The
package does not infer a version from payload content and does not implement
transport, WSDL, WS-Security, or application mapping behavior.

## Documentation

- [Quickstart](docs/QUICKSTART.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Public API reference](docs/API.md)
- [Supported formats and limitations](docs/FORMATS.md)
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

## Development

```sh
go test ./...
```

See [GOAL.md](GOAL.md) for the complete v1 scope and acceptance criteria.
