# wire

[![CI](https://github.com/faustbrian/go-wire/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-wire/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-wire/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-risk_selected-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-optional_local%2Fselected_scheduled-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-wire/v3.svg)](https://pkg.go.dev/github.com/faustbrian/go-wire/v3)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-wire?sort=semver)](https://github.com/faustbrian/go-wire/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`wire` provides explicit, auditable JSON, XML, SOAP, YAML, TOML,
MessagePack, CBOR, and BSON interoperability boundaries with bounded read and
write APIs.

## Status

The latest published stable release is v3.0.1, using
`github.com/faustbrian/go-wire/v3` for the v3 MessagePack admission and
projection contract. Version 2 remains available for consumers not migrating,
but does not include those narrowed controls. See the
[versioned threat model](docs/security-threat-model-v1.md) before adoption.

This source tree prepares the v3.0.2 patch release, fixing lexical SOAP section
capture after length-changing charset conversion. Its dated changelog and
module version identify a release candidate, not completed publication. Adopt
v3.0.2 only after its immutable release tag is published; the installation
command below intentionally selects the already-published v3.0.1 release.

## Requirements

- Go 1.27.0 or later

## Installation

```sh
go get github.com/faustbrian/go-wire/v3@v3.0.1
```

Import format packages explicitly, such as
`github.com/faustbrian/go-wire/v3/jsonwire` or
`github.com/faustbrian/go-wire/v3/soap`. Keep imports within the selected major.

## Quickstart

```go
var response struct {
    Status string `json:"status"`
}

err := jsonwire.DecodeReader(
    strings.NewReader(`{"status":"ok"}`),
    &response,
    jsonwire.DecodeOptions{DisallowUnknownFields: true},
)
if err != nil {
    return err
}
```

Each format exposes explicit options and limits. See the
[quickstart](docs/quickstart.md) for XML namespace validation, SOAP envelopes,
and binary-format examples.

## Package Guarantees

- bounded decode and encode paths for every supported format
- explicit format packages instead of a lossy universal codec
- deterministic output where the format and selected mode permit it
- stable error classification and strict defaults
- documented charset, depth, collection, extension, and data-shape limits
- no HTTP policy, WSDL, schema engine, persistence, or application mapping

The [format matrix](docs/formats.md) is authoritative for supported and
intentionally unsupported behavior. Observable choices at specification and
codec seams are recorded in the
[specification decision register](docs/specification-decisions.md).

## Documentation

Start with the [documentation index](docs/README.md), [quickstart](docs/quickstart.md),
[adoption guide](docs/adoption.md), and [API reference](docs/api.md). Review
[dependencies](docs/dependencies.md), [evidence](docs/evidence.md), and
[hardening](docs/hardening.md) before processing hostile input. Normative
source pins and the conformance update process live in
[specification/README.md](specification/README.md).

AI tools can use [llms.txt](llms.txt) and [llms-full.txt](llms-full.txt).
Support guidance is in [SUPPORT.md](SUPPORT.md), and release history is
maintained in [CHANGELOG.md](CHANGELOG.md).

For ecosystem-wide selection and ownership guidance, see the versioned
[Golib ecosystem index](https://github.com/faustbrian/go-library-tools/tree/v1.4.0/docs/ecosystem)
and its [Integration and data movement family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

## Development

Start with `go test ./...` for ordinary source work, then use `make check` for
bounded local verification. See [contributor guidance](CONTRIBUTING.md#verification)
for focused format fuzzing and selected compatibility, race, integration and
release checks. Mutation is optional local or scheduled assurance, not a PR
prerequisite; `make check-full` retains explicit full assurance.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) and follow the
[code of conduct](CODE_OF_CONDUCT.md). Format-specific behavior and
interoperability tradeoffs must be documented explicitly.

## Security

Report vulnerabilities privately according to [SECURITY.md](SECURITY.md).
Review [docs/security.md](docs/security.md) before decoding untrusted payloads.

## License

`wire` is available under the [MIT License](LICENSE). Attribution and
third-party policy are recorded in [NOTICE](NOTICE) and
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

## Documentation navigation

[Complete documentation map](docs/README.md#complete-documentation-map) · [Current source package map](docs/package-map.md)

Use the [versioned ecosystem entry](https://github.com/faustbrian/go-library-tools/tree/v1.4.0/docs/ecosystem) for package
selection. The [frozen compatibility cohort](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/compatibility-sets.md)
is separate from current package releases and unpublished source changes.
Adopting them does not silently update that cohort.
