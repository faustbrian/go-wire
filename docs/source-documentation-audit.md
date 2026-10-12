# Source documentation audit

## Scope and initial plan

Tier A source documentation only from `1ee9b83aedc4d2e224e071b73311108aa9436489`.
Reproduce with `git ls-files "*.go"`, `git ls-files "*go.mod"` and
`rg "^func (Test|Fuzz|Benchmark|Example)" --glob "*_test.go"`.

One /v3 module and all 82 tracked Go files in 13 source directories are in scope.
The native checker finds 137 missing declaration/field/package comments. Missing
prose is separate from semantic questions that need source/test/spec investigation.

Before contract prose, trace per-format framing and optional presence, borrowed
and owned bytes, strictness/duplicates/canonicalization, target mutation on failure,
callbacks/streaming/reader lifecycle, finite size/depth budgets versus preallocation,
and XML/SOAP/YAML/BSON/MessagePack/CBOR specification interpretations. Audit all
private admission/projection/output/target helpers and every assurance source.

Preserve executable tokens/literals/AST/directives, API/dependency/fixture/authority
pins, decision records/history and original checkout. API mode is apidiff at
api/baseline-v3.txt; no full-prose snapshot is configured. Selected verification
is native formatting/rendering, exact all-source structural proof, native docs
recipe (links/LLM projection/Examples), source/spec checks, changed-content secrets
and ordinary complete final-diff review. No runtime or behavior campaign.

## Complete tracked inventory

| Go file | Initial gaps | Review boundary |
| --- | --- | --- |
| `adversarial_benchmark_test.go` | 0 | BenchmarkRejectAdversarialDecode; BenchmarkRejectCyclicEncode; additional cases/helpers reviewed |
| `api_inventory_test.go` | 0 | TestPublicAPIReferenceInventoriesEveryExport; helpers/fixture boundaries reviewed |
| `automation_test.go` | 0 | TestSharedRepositoryContract; TestSharedDocumentationConventions; additional cases/helpers reviewed |
| `bsonwire/bsonwire.go` | 28 | official byte-backed aliases, iterative nested/scope admission and post-encode policy |
| `bsonwire/bsonwire_test.go` | 0 | TestDecodeFixturePreservesObjectIDDatetimeAndNumericWidth; TestRoundTripPreservesDecimalBinarySubtypeAndRegex; additional cases/helpers reviewed |
| `bsonwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `bsonwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `bsonwire/structure_security_test.go` | 0 | TestCodeWithScopeDuplicatePolicy; TestMalformedNestedStructureRejectedBeforeDecoderCallback; additional cases/helpers reviewed |
| `cborwire/cborwire.go` | 11 | distinct deterministic profiles, tags, finite provider modes and numerical targets |
| `cborwire/cborwire_test.go` | 0 | TestDecodeFixture; TestDecodeRejectsMalformedTrailingAndDuplicateData; additional cases/helpers reviewed |
| `cborwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `cborwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `cborwire/map_key_contract_test.go` | 0 | TestDecodeDynamicMapKeyContracts; TestDecodeTypedTagMapKeysRetainDefaultPolicy; helpers/fixture boundaries reviewed |
| `cmd/semvercheck/main.go` | 1 | release-tool stdout/exit status, no Git operations |
| `cmd/semvercheck/main_test.go` | 0 | TestNextStableVersion; helpers/fixture boundaries reviewed |
| `dependency_differential_test.go` | 0 | TestJSONRejectsInvalidUTF8AcceptedByStandardLibrary; TestMessagePackRejectsDuplicateAcceptedByDependency; additional cases/helpers reviewed |
| `diagnostic_privacy_test.go` | 0 | TestDefaultDiagnosticTextPreservesInspectableCause; TestParserDiagnosticNamesArePrivate; additional cases/helpers reviewed |
| `doc.go` | 0 | canonical format/error boundary and application-owned transport |
| `encoding_limits_test.go` | 0 | TestEveryEncoderEnforcesOutputLimit; TestEncodeWriterDoesNotExceedOutputLimit; additional cases/helpers reviewed |
| `encoding_safety_test.go` | 0 | TestMessagePackCyclicEncodeReturnsError; TestAllEncodePathsRejectCyclicValues; additional cases/helpers reviewed |
| `error.go` | 22 | fixed classification versus sensitive structured/cause fields |
| `error_privacy_test.go` | 0 | TestDecodeErrorsDoNotEchoSensitiveValues; helpers/fixture boundaries reviewed |
| `examples_test.go` | 0 | ExampleDetectFormat; Example; additional cases/helpers reviewed |
| `format.go` | 7 | leading-byte hint; BOM/whitespace and scalar refusal |
| `input_shape_test.go` | 0 | TestEveryDecoderDefinesEmptyWhitespaceTruncatedAndConcatenatedInput; helpers/fixture boundaries reviewed |
| `internal/outputlimit/outputlimit.go` | 0 | partial retained writes, byte alias lifetime and fallback admission |
| `internal/outputlimit/outputlimit_test.go` | 0 | TestBufferEnforcesConfiguredAndFallbackLimits; TestNewRejectsNegativeLimit; helpers/fixture boundaries reviewed |
| `internal/semver/semver.go` | 3 | SemVer syntax versus unchecked native numeric bump |
| `internal/semver/semver_test.go` | 0 | TestValidateTag; TestNextStable; helpers/fixture boundaries reviewed |
| `internal/valuecheck/valuecheck.go` | 0 | active-path recursion, shared acyclic references and callback exclusions |
| `internal/valuecheck/valuecheck_test.go` | 0 | TestValidateRejectsCyclesAndDepth; TestValidateAcceptsAcyclicValues; additional cases/helpers reviewed |
| `jsonwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `jsonwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `jsonwire/jsonwire.go` | 7 | UTF-8, duplicate preflight, float64 decoding and normalization output default |
| `jsonwire/jsonwire_test.go` | 0 | TestDecodeFixture; TestDecodeReaderRejectsUnknownFieldsWhenRequested; additional cases/helpers reviewed |
| `msgpackwire/admission.go` | 0 | abstract work reservation and checked expanded-key arithmetic |
| `msgpackwire/admission_internal_test.go` | 0 | TestAdmissionReservationChecksBeforeMultiplication; TestEmptyArrayProjectionRetainsComponentPolicy; helpers/fixture boundaries reviewed |
| `msgpackwire/admission_test.go` | 0 | TestDecodeAggregateAdmission; TestDecodeAggregateNegativeOptionsBeforeReading; additional cases/helpers reviewed |
| `msgpackwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `msgpackwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `msgpackwire/msgpackwire.go` | 11 | structural/numeric materialization, exact budgets and driver delegation |
| `msgpackwire/msgpackwire_internal_test.go` | 0 | TestDecodeNumericMapRejectsTruncation; TestValidateMessagePackMapBoundaries; additional cases/helpers reviewed |
| `msgpackwire/msgpackwire_test.go` | 0 | TestDecodeFixture; TestDecodeRejectsMalformedTrailingAndUnknownExtensions; additional cases/helpers reviewed |
| `msgpackwire/ordinary_admission_success_internal_test.go` | 0 | TestOrdinaryStructuralMapWithoutBudget; helpers/fixture boundaries reviewed |
| `msgpackwire/ordinary_admission_success_test.go` | 0 | TestOrdinaryMapAdmissionWithLargeWorkAllowance; TestOrdinaryEmptyArrayResetsStruct; helpers/fixture boundaries reviewed |
| `msgpackwire/ordinary_key_weight_internal_test.go` | 0 | TestOrdinaryKeyWeightCheckedAdmission; helpers/fixture boundaries reviewed |
| `msgpackwire/ordinary_owner_internal_test.go` | 0 | TestOrdinaryAdmissionOwners; TestOrdinaryProjectionScalarOwners; additional cases/helpers reviewed |
| `msgpackwire/ordinary_owner_test.go` | 0 | TestOrdinaryProjectionByteArrayDestinations; TestOrdinaryProjectionNestedDestinationRefusal; additional cases/helpers reviewed |
| `msgpackwire/ordinary_projection_cost_internal_test.go` | 0 | TestOrdinaryZeroSizedArrayProjectionCostOverflow; TestOrdinaryZeroSizedArrayProjectionCostControl; helpers/fixture boundaries reviewed |
| `msgpackwire/ordinary_projection_support_internal_test.go` | 0 | TestOrdinaryProjectionDeclinesUnsupportedComponents; TestOrdinaryProjectionPreservesEmptyComponentDriverFallback; additional cases/helpers reviewed |
| `msgpackwire/ordinary_recursive_projection_internal_test.go` | 0 | TestOrdinaryProjectionPreservesRecursiveSourceRefusal; helpers/fixture boundaries reviewed |
| `msgpackwire/projection.go` | 0 | supported target equality, opaque callback/field routing and driver fallback |
| `msgpackwire/projection_test.go` | 0 | TestDecodeRejectsProjectedMapKeys; TestDecodeProjectionPreservesOptionsAndFieldOwnership; additional cases/helpers reviewed |
| `reader_boundary_test.go` | 0 | TestEveryReaderStopsAtOneByteBeyondLimit; helpers/fixture boundaries reviewed |
| `roundtrip_fuzz_test.go` | 0 | FuzzRoundTrip; helpers/fixture boundaries reviewed |
| `soap/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `soap/encode_limit_test.go` | 0 | TestMarshalEnforcesEveryOutputCutoff; TestMarshalFault11EnforcesEveryOutputCutoff; additional cases/helpers reviewed |
| `soap/mutation_resource_guard_unix_test.go` | 0 | TestMain; helpers/fixture boundaries reviewed |
| `soap/soap.go` | 24 | version/cardinality, raw coordinates, lexical faults, shallow fault copy and emission |
| `soap/soap_internal_test.go` | 0 | TestDefaultCharsetReaderUsesParseQuota; TestExceedsLimitHonorsExactBoundary; additional cases/helpers reviewed |
| `soap/soap_test.go` | 0 | TestParseSOAP11EnvelopePreservesRawSections; TestParseReaderSupportsBoundedStreams; additional cases/helpers reviewed |
| `target_boundary_test.go` | 0 | TestEveryDecoderClassifiesInvalidTargets; helpers/fixture boundaries reviewed |
| `target_reuse_test.go` | 0 | TestEveryDecoderSupportsSuccessfulTargetReuse; TestEveryDecoderTreatsFailedTargetAsIndeterminate; helpers/fixture boundaries reviewed |
| `tomlwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `tomlwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `tomlwire/tomlwire.go` | 4 | empty/native document values and post-assignment strict fields |
| `tomlwire/tomlwire_test.go` | 0 | TestDecodeFixturePreservesDatetimeAndNumericTypes; TestDecodeRejectsMalformedDuplicateAndTrailingData; additional cases/helpers reviewed |
| `wire_test.go` | 0 | TestDetectFormat; TestDetectFormatRejectsUnknownAndEmptyPayloads; additional cases/helpers reviewed |
| `xmlwire/charset_security_test.go` | 0 | TestCharsetReaderRejectsLabelsBeforeReading; TestCharsetReaderReadErrorIsPrivateAndPreservesCause; additional cases/helpers reviewed |
| `xmlwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `xmlwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `xmlwire/mutation_resource_guard_unix_test.go` | 0 | TestMain; helpers/fixture boundaries reviewed |
| `xmlwire/xmlwire.go` | 8 | root/depth versus schema/document profile, charset passes and target decoding |
| `xmlwire/xmlwire_internal_test.go` | 0 | TestDefaultCharsetReaderUsesDecodeQuota; TestExactBoundaryPredicates; additional cases/helpers reviewed |
| `xmlwire/xmlwire_test.go` | 0 | TestDecodeNamespaceAwareFixture; TestRootReturnsResolvedNamespace; additional cases/helpers reviewed |
| `yamlwire/doc.go` | 0 | canonical package ownership, normal entry and pinned format profile |
| `yamlwire/encode_dependency_test.go` | 0 | TestEncodeClassifiesDumperConstructionFailure; helpers/fixture boundaries reviewed |
| `yamlwire/encode_limit_test.go` | 0 | TestEncodeEnforcesOutputLimit; helpers/fixture boundaries reviewed |
| `yamlwire/scalar_integrity_test.go` | 0 | TestEncodePreservesScalarSyntaxAndBlockContents; TestEncodeWriterPreservesScalarSyntaxAndFinalQuota; additional cases/helpers reviewed |
| `yamlwire/yamlwire.go` | 11 | node preflight, pinned emitter repair, scratch/final quotas and callback occurrences |
| `yamlwire/yamlwire_internal_test.go` | 0 | TestBoundaryPredicates; TestAliasLimitPredicates; additional cases/helpers reviewed |
| `yamlwire/yamlwire_test.go` | 0 | TestDecodeFixture; TestDecodeReaderRejectsUnknownFieldsWhenRequested; additional cases/helpers reviewed |

## Investigated constraints and open questions

- All 25 implementation/helper files and 57 assurance files are audited. The
  shared target/reuse/reader/output/privacy tests trace complete input, one-byte
  refusal, negative/exact limits, zero-progress writes and indeterminate targets.
  Every codec fixture/helper/seed and platform guard is included, not executed as
  broad qualification. Nine fuzz targets and 18 benchmarks retain their actual
  input partitions, setup exclusions and comparison limitations.
- Reader facades buffer one complete unit, through EOF or one detection byte
  beyond quota (except the MaxInt64 overflow boundary). They borrow and never
  close readers; there is no context/deadline API. Custom codecs/readers/writers
  must cooperate, terminate and receive application admission. Typed-nil IO
  collaborators are not generally screened by literal interface nil checks.
- All typed encoders perform active-path cycle/depth admission with per-call state;
  shared acyclic references are valid. Unexported fields are skipped and callbacks
  are not executed in graph preflight. Retained-output quotas do not cap codec
  intermediate buffering, caller graphs, allocations or elapsed callback time.
  Encoding errors publish no returned payload; writer failures can expose a prefix.
- JSON rejects invalid UTF-8, offers duplicate preflight, otherwise retains the
  stdlib's replacement/merge semantics. Normalize preserves number lexemes but
  re-encodes using the default 1 MiB output limit, independently of input override.
  Neither deterministic output nor normalization is an RFC 8785 signing profile.
- XML uses two independent charset/token/target passes, one element root and depth
  admission. Root ignores ExpectedRoot and non-element character data outside the
  root; it is not full document/schema/DTD validation. No declared-entity fetching
  is installed. Charset conversion limits raw bytes, can expand UTF-8 up to three
  times, and custom callbacks retain output/work responsibility.
- SOAP parses strict version namespaces and Header-before-Body cardinality; valid
  Fault publishes both Envelope and FaultError. Raw accessors copy bytes, exported
  fields are mutable, and FaultError's shallow copy aliases nested fault slices.
  Fault codes/subcodes remain lexical rather than resolved QNames. DecodeBody can
  mutate its target before later cardinality failure. Raw Marshal validation is
  neither full Parse depth policy, single-child enforcement nor schema processing.
  Source-offset slicing assumes raw/decoded byte-coordinate agreement. Bounded
  recover-only observation of a windows-1252 envelope containing 32 byte-0x80 body
  characters confirms incorrect BodyXML with spare capacity and a panic with
  exact input capacity. The observed shape is SOAP 1.1 Envelope/Body/v with no
  Header, a windows-1252 XML declaration and default ParseOptions; make/copy into
  a len-equals-cap slice exposes the panic at parseBody/innerEnd. Matching-
  declaration UTF-8 control succeeds. No runtime repair is included; callback
  charset availability is not raw-capture safety.
- YAML preflights nodes then loads targets under pinned v4 policy and optional
  positive alias/depth plugins. Emitter-only scalar repair preserves authored
  logical breaks/quoted contexts without replaying callbacks. Scratch quota can be
  twice final quota, with checked native arithmetic before normalized allocation;
  this is not an arbitrary-input YAML parser or universal canonicalizer.
- MessagePack raw admission and numeric/projected uniqueness have distinct owners.
  Abstract comparison reservation precedes generic materialization; unsupported
  custom/struct/interface/inline routing retains driver-owned boundaries. Current
  precise admission/projection rationale is preserved. Driver map sorting is shape-
  dependent, compact modes alter representation, and custom registration owns
  extension/key behavior. Integer/float narrowing checks are not a blanket exact
  integer-to-float guarantee: pinned driver signed-bit conversion remains visible
  in the existing projection owner controls. No final-driver heap/context claim.
- CBOR profiles are separately named RFC 7049 canonical, RFC 8949 Core Deterministic
  and CTAP 2.2. Duplicate policy, tags/indefinite admission and finite provider
  nesting/count limits remain distinct from output profile and byte ceilings.
- BSON aliases preserve official driver types/byte ownership, not independent
  wrappers. Complete length/terminator and iterative document/array/scope validation
  precede target callbacks, with 100 nesting levels below root. Post-encode
  validation cannot bound driver buffering. Ordered D/struct/Raw and unordered M
  differ intentionally; duplicate/truncation/width options change interoperability.
- TOML accepts empty documents, native dates/times and schema-specific values.
  Unknown-field refusal follows target decode, so it is not transactional.
- Internal release helpers validate SemVer syntax, not Git/release permission.
  NextStable ignores numeric conversion errors and unchecked increments; enormous
  components are not a validated release-input contract. No runtime repair here.

Pinned sources remain in specification/manifest.tsv: JSON RFC 8259, XML 1.0 Fifth
Edition and Namespaces Third Edition, SOAP 1.1/1.2 editions, YAML 1.2.2, TOML 1.1.0,
MessagePack commit, CBOR RFCs/CTAP 2.2, BSON 1.1, and historical Go 1.26.6 authority
snapshots. Runtime verification uses the repository Go 1.27.2 pin; historical pins
are not rewritten. No generated Go, generation/deprecation or unfinished source
markers were found. The two Unix-only mutation guards keep their build directives.

Future behavior questions are not prose gaps: transactional targets/writes,
context-aware callback/reader APIs, strict preallocation/time quotas, general
MessagePack projected-key coverage, XML document-profile tightening, SOAP charset
source coordinates and checked release-version arithmetic require separate scope.
No new canonicalization, conformance, interoperability, benchmark or release claim.

## Objective Enforcement

The checksum-pinned public Tools v2.4.0 verifier now supplies
`make source-docs-check` as a prerequisite of the existing local and CI Make
routes. Ordinary shared CI selects strict source-documentation checking and
requires its result. Native prerequisites, operation budgets, security and
release controls are retained. Hosted exact-head evidence is recorded by the
coordination ledger; configuration alone does not establish green delivery.
