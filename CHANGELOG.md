# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Adopt the `go-library-tools` v1.3.0 schema-v2 cohesion contract and local
  `make cohesion` gate without changing wire APIs or runtime behavior.
- Pin reusable CI to the immutable v1.3.0 workflow and enforce cohesion
  metadata in the repository's required CI contract.

- Adopt checksum-pinned `go-library-tools` v1.2.0 and its immutable workflow
  so CI executes specification governance while keeping format-specific
  conformance, interoperability, mutation, and benchmark evidence in this
  repository.

### Documentation

- Record the reviewed Go releases-feed representation change as
  behavior-neutral for the unchanged immutable Go 1.26.6 source contracts.

- Replace commit-pinned installation and pre-v1 guidance with the published
  v1.0.0 release, exact Go 1.26.6 minimum, stable compatibility contract, and
  a package-level compiler-checked example.

- Record the reviewed Go release-feed change through Go 1.27.1 as
  behavior-neutral for the pinned Go 1.26.6 JSON, XML, errors, and language
  contracts; Go 1.27 adoption remains a separate decision review.

- Record the reviewed FIDO Alliance feed change as behavior-neutral for the
  pinned CTAP 2.2 deterministic-CBOR profile; replacing the general-news feed
  with a specification-specific release authority remains future maintenance.

- Record reported RFC 9110 Errata ID 9162 as behavior-neutral because HTTP
  field combination remains caller-owned transport policy; re-review it if the
  erratum becomes verified or the package ownership boundary changes.

- Publish the module's family, capabilities, ownership, lifecycle, supported
  environments, package selection, and delivery status, and link the README to
  the immutable v1.3.0 ecosystem index and family guidance.

- Make the [specification decision register](docs/specification-decisions.md)
  machine-auditable with exact source and change-authority monitoring,
  attributable conformance evidence, classified maintained-peer results, and
  durable decision history.

  - WIRE-DEC-001 sha256:f1fc52ef8464874e7b8cebcae98161f7c9296128e9a8913fcf918b97fe2f0728
  - WIRE-DEC-002 sha256:f4dee326c94a5de595cbd5e0aea03414fd8ec444be92fd1ca25adcaef4a46f6d
  - WIRE-DEC-003 sha256:1dd67beed03dbeb91bdef10c77f6f8c21cd0c8326720861437f9ce1ae179ee98
  - WIRE-DEC-004 sha256:162b1cb67101d82b7732701b50193d49f2702908450227ecae9068777b5c91f5
  - WIRE-DEC-005 sha256:74528fbf1f9dae53b31e20152c679fe0614de1d4a3f45c1fa19179b6a99cb620
  - WIRE-DEC-006 sha256:a283308985c948563734b7f69e182e97ca844068e43e1cc9f5b09fcceb419e98
  - WIRE-DEC-007 sha256:c781bb6f45e052079ba1b1dbdb37b3184606c72abae012848ab23a1d7cc2d73c
  - WIRE-DEC-008 sha256:c5ce0457d9c8f828dfebef5c5d4467760836fcd58bb2b4b6291fe4bc76df4953
  - WIRE-DEC-009 sha256:3739785c9c34062134e2cbe7d38b974da151e1711ced5e16119184b7c0060e91
  - WIRE-DEC-010 sha256:57459c19b8ef283a3891ef4bc4d44efa1f1f4d211af4a9e74a533721c08f03f9
  - WIRE-DEC-011 sha256:78340a1b77fad0a94fdd12875dd851446ff46fea84d92f4eb1766f9fb6637909
  - WIRE-DEC-012 sha256:bbf48898109b852fe1b5784511c0bb5aa308162ceacd10b8777023fa391f9491
  - WIRE-DEC-013 sha256:4a7375048cd8a95c1b37ecaae2e5d21c7dbb97b5e95c544d12fd2db1ccb29b7b
  - WIRE-DEC-014 sha256:26d5aafc3d7e2bbf9304d04d9c0e1bf435968260b18dd05043fef4d1666596bd
  - WIRE-DEC-015 sha256:a78c203aaf7314a61baa093d7e80e87171e8695c570f7e0b443b4ddda227be66
  - WIRE-DEC-016 sha256:977cc72dfe7abd290dff31fa27f8150c26c0c27b4704e31427aabca3e7dffd52

- Remove the archived monorepo documentation link; package guidance remains in
  the repository-owned documentation.

## [1.0.0] - 2026-08-25

### Changed

- Validate action pinning from the standalone repository root and leave
  repository-foundation policy to the authoritative repository contract.

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Replace obsolete standalone-repository links and workflow claims with
  monorepo-canonical targets and current release guidance.

- Link the package README to the repository-wide Golib documentation portal.

### Compatibility

- Added a pinned module export baseline so incompatible public API changes
  fail the canonical repository gate.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-wire` identity while preserving its documented API and behavior.
- Hardened codec boundary coverage with exact byte, depth, numeric, charset,
  alias, fragment, and SOAP fault assertions, and bounded mutation execution
  for malformed XML and SOAP parser mutants.
- Link the conformance source matrix directly to the canonical specification
  decision register.
- Added the `GO-SAFETY-1` ownership, concurrency, race, fuzz, resource, and
  benchmark standard with an executable `make safety` gate.
- Moved AI planning and hardening briefs into `.ai/` and clarified the
  separate purposes of project and third-party notice files.

### Added

- A specification decision register, pinned normative-source manifest, and
  executable conformance gate covering the observable JSON, XML, SOAP, YAML,
  TOML, MessagePack, CBOR, and BSON policy choices.
- Opt-in recursive duplicate-name rejection for bounded JSON decoding before
  the destination value can be mutated.
- A standardized OSS repository skeleton covering policy, documentation,
  legal notices, Go tooling, pinned CI, security, and release automation.
- CI resolves the latest supported Go 1.25 patch while `go.mod` declares the
  portable Go 1.25 minimum.
- Evidence-driven audit and hardening goal covering every supported format,
  parser resource safety, codec dependencies, and read/write boundaries.
- A shared `wire.Error` model with parse, validation, unsupported-format,
  envelope, and SOAP-fault classifications.
- Size-limit, invalid-target, and encoding classifications shared by every
  format package.
- Explicit JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON format
  identifiers.
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
- Bounded YAML read/write APIs with duplicate-key and multi-document policy,
  alias and merge controls, expansion limits, deterministic output, fixtures,
  fuzzing, and benchmarks.
- Complete TOML document read/write APIs with strict-field metadata, native
  datetime handling, checked numeric conversion, deterministic output,
  fixtures, fuzzing, and benchmarks.
- MessagePack read/write APIs with exact one-object enforcement, map-key and
  extension policy, checked numeric widths, timestamp support, deterministic
  map encoding, fixtures, fuzzing, and benchmarks.
- Canonical, Core Deterministic, and CTAP2 CBOR read/write profiles with tag,
  indefinite-length, duplicate-key, and resource-limit policy, plus fixtures,
  fuzzing, and benchmarks.
- BSON document read/write APIs with recursive duplicate-key rejection, exact
  length validation, ordered and raw document support, ObjectID and datetime
  aliases, checked numeric conversion, fixtures, fuzzing, and benchmarks.
- Pinned, reviewed YAML, TOML, MessagePack, CBOR, and BSON codec dependencies
  with documented maintenance, security, and residual-risk rationale.
- Compile-checked round-trip examples and full API, format, migration,
  adoption, security, troubleshooting, and dependency documentation for all
  eight supported formats.
- A distinct `wire.ErrWrite` classification for destination failures.
- Consistent repository automation for generated portable AI documentation,
  dependency review, and guarded semantic release commands.
- Configurable output byte limits for every JSON, XML, SOAP, YAML, TOML,
  MessagePack, CBOR, and BSON byte and writer encoding path. Zero selects the
  safe 1 MiB default.
- Size-bounded SOAP raw-envelope and fault APIs through `MarshalOptions`,
  `MarshalWithOptions`, `MarshalWriterWithOptions`,
  `MarshalFaultWithOptions`, and `MarshalFaultWriterWithOptions`.

### Changed

- The minimum supported Go version is Go 1.25.0; later Go 1.25 patch
  releases and newer language versions are supported.
- Fuzz, benchmark, release, and documentation automation now covers YAML,
  TOML, MessagePack, CBOR, and BSON alongside JSON, XML, and SOAP.
- Corrected codec troubleshooting and migration guidance to use the exported
  option and profile names, describe YAML's actual safe defaults, distinguish
  legacy size classifications, and document BSON double truncation as
  explicitly lossy.
- MessagePack decoding now enforces configurable safe defaults of 32 nesting
  levels, 131,072 array elements, and 65,536 map pairs. Structural limit
  failures are classified as `wire.ErrSizeLimit`.
- MessagePack decoding now rejects duplicate map keys recursively by default,
  with explicit last-key-wins compatibility through `AllowDuplicateKeys`.
- BSON now re-exports the official array, raw value, Decimal128, binary, regex,
  timestamp, JavaScript, scoped code, sentinel, pointer, and symbol types.
- Published the evidence-driven hardening report, threat model, per-format
  conformance matrix, allocation policy, dependency residual risks, and
  pre-v1 semantic-version recommendation.
- Writer APIs now complete encoding within the configured output quota before
  writing, so an encode limit failure cannot emit a partial destination
  payload.

### Fixed

- Keep module-archive tests scoped to files shipped with the wire module;
  repository-root workflow policy remains owned by the root verification gate.
- Bound fuzz-smoke concurrency to avoid deadline flakes on high-core hosts.
- MessagePack now performs allocation-safe structural preflight for impossible
  collection lengths and composite map keys. Numeric preflight rejects
  narrowing overflow in typed maps, array-encoded structs, and automatically
  inlined embedded fields before the decoder can apply a lossy Go conversion.
- MessagePack structural validation now stops recursive traversal and compact
  collection allocation amplification at explicit caller-configurable limits.
- All typed encoders now reject cyclic values and nesting beyond 1,000
  traversed levels before recursive codecs can exhaust the process stack.
- MessagePack numeric preflight now also prevents duplicate keys from being
  silently overwritten before assignment.
- BSON documentation and tests now prove Decimal128, binary subtype, and regex
  interoperability instead of claiming it from an incomplete alias set.
- JSON decoding and normalization now reject invalid UTF-8 as a parse failure
  instead of silently accepting replacement characters.
- XML tests now prove directive handling, non-expansion of declared entities,
  and embedded-NUL rejection.
- SOAP tests now prove fault text and language attributes cannot inject XML
  while validated raw Detail fragments remain explicit.
- Cross-format writer regressions now prove zero-progress destinations are
  rejected as `wire.ErrWrite`, including all three SOAP writer families.
- CBOR tests now lock simple-value, tagged-bignum, and preferred integer and
  shortest-float behavior to the selected deterministic profile.
- TOML tests now lock dotted-key conflicts, special floats, and arrays of
  tables to the documented dependency behavior.
- YAML tests now lock custom tags, core-schema implicit values, and typed
  non-JSON map keys to the documented compatibility boundary.
- YAML now classifies the dependency's built-in excessive-alias and maximum
  nesting protections as `wire.ErrSizeLimit`, matching caller-configured
  resource limits.
- YAML merge-key rejection now follows the parsed merge tag without rejecting
  an ordinary quoted `<<` scalar value.
- Every encoder now rejects output beyond its configured byte limit as
  `wire.ErrSizeLimit` instead of allowing unbounded output buffering.
- SOAP raw header, body, and fault Detail fragments that cannot fit the output
  quota are rejected before XML validation allocates a wrapper copy.
- Package-local boundary regressions now prove exact, undersized, negative,
  and dependency-construction behavior while retaining 100% production
  statement coverage.
- XML and SOAP parsing now enforce a configurable 1,000-element default token
  depth and classify excess nesting as `wire.ErrSizeLimit`.
- A cross-format round-trip fuzzer now proves writer/reader semantic parity for
  bounded strings, signed integers, and booleans across all eight formats.
- YAML block scalars now use an explicit indentation indicator, preventing the
  YAML v4 emitter from producing tab-leading multiline output that its own
  parser rejects.
- JSON, XML, and SOAP input byte limits now use `wire.ErrSizeLimit`, and their
  invalid decode targets now use `wire.ErrTarget`, matching every other format.
- Hostile-reader regressions prove every streaming decoder reads no more than
  `MaxBytes + 1` bytes before rejecting an oversized source.
- Cross-format target tests prove successful reuse and record partial mutation
  on conversion failure; adversarial benchmarks cover hostile decode shapes
  and cyclic encodes with allocation counts.
- Decoder fuzz corpora now seed empty/whitespace input, deep nesting,
  duplicates, invalid encodings and numbers, multiple documents, aliases,
  tags/extensions, and malformed binary length fields as applicable.
- Dependency-differential tests lock the wrapper's stricter invalid-UTF-8 and
  duplicate-key policies and the YAML block-indentation repair against pinned
  codec behavior.
- A compiler-derived public API inventory test and audit evidence ledger map
  every exported option, boundary requirement, format hazard, fuzzer, and
  benchmark to traceable tests.
- An all-format input-shape matrix locks empty, whitespace-only, truncated, and
  concatenated behavior, including binary whitespace-as-data semantics.
- The public API inventory parses production files explicitly without the
  deprecated Go 1.25 `parser.ParseDir` helper.
- Cross-format malformed-input regressions ensure classified decoder errors do
  not echo sensitive values from the payload, and failed-target coverage avoids
  promising a dependency-specific partial-assignment order.
- The final hardening verdict records Go 1.25 compatibility, symmetric bounded
  APIs, current direct codec dependencies, and the remaining upstream risks.

[Unreleased]: https://github.com/faustbrian/go-wire/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/faustbrian/go-wire/releases/tag/v1.0.0
