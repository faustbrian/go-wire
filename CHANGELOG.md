# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- A shared `wire.Error` model with parse, validation, unsupported-format,
  envelope, and SOAP-fault classifications.
- Explicit JSON, XML, and SOAP format identifiers.
- Opt-in JSON/XML format detection based on the first significant byte.
- Bounded JSON byte-slice and reader decoding with optional unknown-field
  rejection and single-value enforcement.
- Deterministic JSON encoding with configurable indentation and HTML escaping.
- Explicit JSON normalization that strips a UTF-8 BOM, compacts whitespace,
  orders object keys, and preserves number lexemes.
- JSON fixtures, malformed-input regression tests, fuzz seeds, and parse and
  encode benchmarks.
- Bounded strict XML decoding, opt-in non-strict recovery, exact
  namespace-aware root validation, and resolved-root inspection.
- Deterministic XML encoding with optional declaration and indentation.
- Built-in, explicit charset conversion for UTF-8, US-ASCII, ISO-8859-1, and
  Windows-1252, including rejection of invalid or undefined bytes.
- Namespace, malformed-document, charset, and trailing-root XML fixtures and
  regressions, plus fuzz seeds and parse and encode benchmarks.
- Bounded SOAP 1.1 and 1.2 envelope parsing with exact raw envelope, header,
  and body access and namespace-preserving body decoding.
- Typed SOAP 1.1 and 1.2 fault extraction, localized SOAP 1.2 reasons,
  subcodes, details, and `errors.Is` classification.
- Deterministic envelope and fault serialization from validated raw fragments.
- SOAP envelope, fault, malformed-input, structural-regression, fuzz, and
  benchmark coverage.
- Compile-checked examples and adoption documentation covering quickstart,
  architecture, the complete public API, supported formats, behavioral
  guarantees, limitations, and rollout guidance.
- End-to-end JSON, XML, and SOAP examples, a scenario cookbook, FAQ, and
  error-focused troubleshooting guide.
- Migration, versioning, release, contribution, security, conduct, and roadmap
  documentation for open-source operation.
- Reproducible local formatting, vet, lint, race-test, 100% coverage, fuzz,
  benchmark, documentation-link, and vulnerability quality gates.
- GitHub Actions for the Go 1.25 and 1.26 build/race-test matrix,
  formatting, static analysis, lint, exact coverage, documentation, fuzz smoke,
  benchmark smoke, and vulnerability scanning.
- Scheduled extended fuzzing and benchmark baselines, Dependabot configuration,
  and an interoperability-focused pull request template.
- Tagged release automation that validates SemVer tags and changelog entries,
  reruns quality and security gates, creates a deterministic source archive and
  checksum, and publishes a GitHub Release with extracted notes.
- Explicit method-level documentation for every exported error API.
- JSON and XML writer APIs, typed SOAP header/body encoding, and SOAP envelope
  and fault writer APIs for symmetric input and output handling.
- A distinct `wire.ErrWrite` classification for destination failures.
- Consistent repository automation for generated portable AI documentation,
  dependency review, and guarded semantic release commands.

### Changed

- The minimum supported Go version is 1.25.12.

[Unreleased]: https://github.com/faustbrian/go-wire/compare/v0.0.0...HEAD
