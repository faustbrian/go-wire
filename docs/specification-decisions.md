# Wire Specification Decisions

This register records observable choices where the supported wire-format
specifications, Go codecs, dependency codecs, and package policy permit
different outcomes. Each format remains an explicit package; this module does
not claim one lossless data model across all eight formats.

Each resolved entry names executable evidence. Changing one requires
compatibility, security, resource, API, conformance, dependency, and changelog
review. Superseded decisions remain linked from their replacements.

## WIRE-DEC-001: Specification editions and codec delegation

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-001","title":"Specification editions and codec delegation","status":"resolved","owner":"wire maintainers","classification":"implementation-defined behavior","decision_scope":"application-policy","specification":"Go 1.26.6 encoding/json contract","version":"Go 1.26.6","source_authority":"go-json-source","section":"src/encoding/json and src/encoding/xml","requirement_strength":"not specified","issue":"The wire specifications define distinct data models while maintained Go codecs expose overlapping subsets and defaults.","interpretations":["Reimplement every grammar.","Expose dependency defaults unchanged.","Wrap pinned codecs with explicit package policy."],"peer_behavior":"Maintained Go and dependency codecs are compared where wrapper policy deliberately differs.","selected_behavior":"Delegate syntax and Go value mapping to the pinned maintained codecs while keeping limits, options, and error classification in this package.","rationale":"Delegation preserves maintained format implementations without claiming one lossless cross-format model.","security_consequences":"Package preflight and bounded wrapper policy remain authoritative at hostile input boundaries.","resource_consequences":"Finite package limits constrain codec work and allocation.","compatibility_consequences":"Codec upgrades require decision and differential review.","wire_consequences":"Each format retains its own data model and codec-defined representation within documented package policy.","executable_evidence":["TestSharedRepositoryContract","TestCodecDocumentationUsesRealAPIsAndSemantics"],"fixture_evidence":[],"fuzz_evidence":["FuzzRoundTrip"],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["All public format packages"],"documentation":["docs/specification-decisions.md","specification/README.md"],"upstream_status":"Source editions and maintained codec versions are pinned and monitored separately.","reconsider_when":"Any source edition, codec version, or supported format changes."}
```

Authority URL: https://api.github.com/repos/golang/go/contents/src/encoding/json?ref=go1.26.6

Additional authoritative source: `{"id":"go-xml-source","version":"Go 1.26.6","url":"https://api.github.com/repos/golang/go/contents/src/encoding/xml?ref=go1.26.6","specifications":["Go 1.26.6 encoding/xml contract"]}`

Additional authoritative source: `{"id":"rfc8259-source","version":"RFC 8259","url":"https://www.rfc-editor.org/rfc/rfc8259.txt","specifications":["RFC 8259 JSON"]}`

Additional authoritative source: `{"id":"xml10-source","version":"XML 1.0 Fifth Edition","url":"https://www.w3.org/TR/2008/REC-xml-20081126/","specifications":["XML 1.0 Fifth Edition and Namespaces in XML 1.0 Third Edition"]}`

Additional authoritative source: `{"id":"xml-names-source","version":"Namespaces in XML 1.0 Third Edition","url":"https://www.w3.org/TR/2009/REC-xml-names-20091208/","specifications":["XML 1.0 Fifth Edition and Namespaces in XML 1.0 Third Edition"]}`

Additional authoritative source: `{"id":"soap11-source","version":"SOAP 1.1","url":"https://www.w3.org/TR/2000/NOTE-SOAP-20000508/","specifications":["SOAP 1.1 and SOAP 1.2 Part 1 Second Edition"]}`

Additional authoritative source: `{"id":"soap12-source","version":"SOAP 1.2 Part 1 Second Edition","url":"https://www.w3.org/TR/2007/REC-soap12-part1-20070427/","specifications":["SOAP 1.1 and SOAP 1.2 Part 1 Second Edition"]}`

Additional authoritative source: `{"id":"yaml122-source","version":"YAML 1.2.2","url":"https://raw.githubusercontent.com/yaml/yaml-spec/80c6bde43887cf83424defb5c45c2485464659b9/spec/1.2.2/spec.md","specifications":["YAML 1.2.2"]}`

Additional authoritative source: `{"id":"toml110-source","version":"TOML 1.1.0","url":"https://raw.githubusercontent.com/toml-lang/toml/bcbbd1c1f03473ffe97b8bf26a0fc945efe2b4a1/toml.md","specifications":["TOML 1.1.0"]}`

Additional authoritative source: `{"id":"msgpack-source","version":"MessagePack commit 8aa09e2a6a91","url":"https://raw.githubusercontent.com/msgpack/msgpack/8aa09e2a6a9180a49fc62ecfefe149f063cc5e4b/spec.md","specifications":["MessagePack format at 8aa09e2a6a91"]}`

Additional authoritative source: `{"id":"rfc7049-source","version":"RFC 7049","url":"https://www.rfc-editor.org/rfc/rfc7049.txt","specifications":["RFC 7049 and RFC 8949 CBOR"]}`

Additional authoritative source: `{"id":"rfc8949-source","version":"RFC 8949","url":"https://www.rfc-editor.org/rfc/rfc8949.txt","specifications":["RFC 7049 and RFC 8949 CBOR"]}`

Additional authoritative source: `{"id":"ctap22-source","version":"CTAP 2.2 Proposed Standard 2025-07-14","url":"https://fidoalliance.org/specs/fido-v2.2-ps-20250714/fido-client-to-authenticator-protocol-v2.2-ps-20250714.pdf","specifications":["CTAP 2.2 deterministic CBOR profile"]}`

Additional authoritative source: `{"id":"bson11-source","version":"BSON 1.1","url":"https://bsonspec.org/spec.html","specifications":["BSON 1.1"]}`

</details>

**Authoritative reference:** [Go 1.26.6 encoding source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.6:src/encoding/).

- **Status, owner, and classification:** `resolved`; wire maintainers;
  normative-source and interoperability policy.
- **Source and issue:** The pinned
  [source manifest](../specification/manifest.tsv) identifies the exact JSON,
  XML 1.0, SOAP 1.1/1.2, YAML 1.2.2, TOML 1.1.0, MessagePack, CBOR, CTAP2, and
  BSON 1.1 sources. They define distinct data models, while Go and third-party
  codecs implement overlapping but not identical subsets and defaults.
- **Interpretations and peer behavior:** Reimplement every grammar, expose
  dependency defaults unchanged, normalize all formats through one generic
  tree, or wrap reviewed codecs with explicit package policy. Codec behavior
  differs on duplicate names, numeric conversion, tags, aliases, and output
  ordering.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  JSON and XML delegate syntax and Go
  value mapping to Go 1.26.6; SOAP builds on that XML boundary. YAML, TOML,
  MessagePack, CBOR, and BSON use the exact modules in `go.mod`, with package
  preflight, limits, options, and error classification defining the public
  contract. Dependency behavior is never promoted to a broader conformance
  claim without package evidence.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestSharedRepositoryContract`, `TestCodecDocumentationUsesRealAPIsAndSemantics`,
  and `TestPublicAPIReferenceInventoriesEveryExport` cover all format packages
  and documentation. Reconsider every source or codec upgrade and any new
  format.

## WIRE-DEC-002: Complete input units and destination mutation

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-002","title":"Complete input units and destination mutation","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"defensive","specification":"RFC 8259 JSON","version":"RFC 8259","source_authority":"rfc8259-source","section":"Sections 2 and 3","requirement_strength":"not specified","issue":"Format documents do not define one uniform Go decoder policy for empty input, concatenated units, invalid targets, or partial destination mutation.","interpretations":["Decode the first value.","Decode streams implicitly.","Require exactly one complete unit with explicit stream APIs."],"peer_behavior":"Maintained-peer destination mutation after errors has not been assessed as a portable contract.","selected_behavior":"Require one complete non-empty unit except for TOML's valid empty document, and treat destinations as indeterminate after decode errors.","rationale":"A single-value API must not silently become a stream API or promise rollback it cannot provide.","security_consequences":"Trailing or concatenated attacker-controlled data cannot be ignored.","resource_consequences":"Readers stop at a bounded detection byte beyond the configured limit.","compatibility_consequences":"Callers needing streams must select the explicit YAML multi-document option or a future streaming API.","wire_consequences":"Successful decode consumes exactly the documented format unit.","executable_evidence":["TestEveryDecoderDefinesEmptyWhitespaceTruncatedAndConcatenatedInput","TestEveryDecoderClassifiesInvalidTargets","TestEveryDecoderSupportsSuccessfulTargetReuse","TestEveryDecoderTreatsFailedTargetAsIndeterminate"],"fixture_evidence":[],"fuzz_evidence":[],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["All Decode and DecodeReader functions"],"documentation":["docs/specification-decisions.md"],"upstream_status":"No cross-format standard defines destination rollback semantics.","reconsider_when":"A separately owned streaming or transactional decode API is introduced."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc8259.txt

</details>

**Authoritative reference:** [RFC 8259](https://www.rfc-editor.org/rfc/rfc8259.html).

- **Status, owner, and classification:** `resolved`; maintainers; defensive
  application-boundary policy.
- **Source and issue:** Format specifications define streams, documents, or
  data items, while decoder APIs may accept prefixes, multiple units, empty
  documents, or mutate destinations before a later failure.
- **Interpretations and peer behavior:** Decode the first value, decode a
  stream implicitly, require exactly one unit, or expose per-format behavior.
  Codec destinations commonly become partially assigned on errors.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  JSON, XML, SOAP, YAML, MessagePack,
  CBOR, and BSON require exactly one non-empty complete value, document,
  envelope, or item. TOML permits its specified empty document. YAML multi-
  document input is opt-in and requires a slice destination. All decoders
  reject nil and non-pointer targets before reading; successful reuse is
  supported, but a destination is explicitly indeterminate after any decode
  error.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestEveryDecoderDefinesEmptyWhitespaceTruncatedAndConcatenatedInput`,
  `TestEveryDecoderClassifiesInvalidTargets`,
  `TestEveryDecoderSupportsSuccessfulTargetReuse`, and
  `TestEveryDecoderTreatsFailedTargetAsIndeterminate` cover every decode API.
  Reconsider only with an explicit streaming API that has separate ownership.

## WIRE-DEC-003: Input, structural, graph, and output limits

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-003","title":"Input, structural, graph, and output limits","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"defensive","specification":"RFC 7049 and RFC 8949 CBOR","version":"RFC 8949","source_authority":"rfc8949-source","section":"Sections 3 through 5","requirement_strength":"not specified","issue":"The wire formats permit values larger or deeper than a general service can safely process and do not select package resource budgets.","interpretations":["Leave limits to callers or codecs.","Impose only a byte limit.","Enforce finite format-aware input, structure, graph, and output limits."],"peer_behavior":"Maintained codec defaults vary and are not treated as a portable shared limit contract.","selected_behavior":"Use finite defaults, validate negative limits, preflight format structure, and complete output within quota before writing.","rationale":"Bounded work and all-or-error output are required at an untrusted serialization boundary.","security_consequences":"Forged lengths, deep graphs, and oversized output are rejected before disproportionate work or partial emission.","resource_consequences":"The default input and output budget is 1 MiB with explicit format-aware structural limits.","compatibility_consequences":"Callers may select stricter or documented larger limits without changing format semantics.","wire_consequences":"Limit failures emit no partial encoded document.","executable_evidence":["TestEveryReaderStopsAtOneByteBeyondLimit","TestEveryEncoderHonorsExactNegativeAndMaximumOutputLimits","TestEncodeWriterDoesNotExceedOutputLimit","TestForgedBinaryLengthsHaveBoundedAllocationCounts"],"fixture_evidence":[],"fuzz_evidence":[],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["All bounded decode and encode APIs"],"documentation":["docs/specification-decisions.md","docs/security.md"],"upstream_status":"Resource budgets remain package policy rather than CBOR or cross-format requirements.","reconsider_when":"Measured service constraints justify a compatibility-reviewed default change."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc8949.txt

</details>

**Authoritative reference:** [RFC 8949](https://www.rfc-editor.org/rfc/rfc8949.html).

- **Status, owner, and classification:** `resolved`; maintainers; defensive
  resource policy.
- **Source and issue:** The format specifications permit documents and values
  far larger or deeper than a general service can safely process, and encoded
  length fields can request disproportionate allocation.
- **Interpretations and peer behavior:** Leave limits to callers or codecs,
  impose one byte limit only, stream unbounded output, or enforce finite
  format-aware limits. Codec defaults and allocation behavior vary.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  Zero selects a 1 MiB input and
  output default. Negative limits fail validation. Readers consume at most one
  detection byte beyond the limit, with `MaxInt64` handled without overflow.
  XML/SOAP depth, YAML alias/depth, MessagePack depth/collection, and CBOR
  dependency limits remain explicit. Every typed or raw encoder completes
  within a quota before writing, so limit failures emit no partial output.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestEveryReaderStopsAtOneByteBeyondLimit`,
  `TestEveryEncoderHonorsExactNegativeAndMaximumOutputLimits`,
  `TestEncodeWriterDoesNotExceedOutputLimit`, and
  `TestForgedBinaryLengthsHaveBoundedAllocationCounts` cover all formats.
  Reconsider defaults only with compatibility and measured resource evidence.

## WIRE-DEC-004: Conservative format detection

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-004","title":"Conservative format detection","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"application-policy","specification":"RFC 9110 HTTP Semantics","version":"RFC 9110","source_authority":"rfc9110-source","section":"Section 8.3 media types","requirement_strength":"not specified","issue":"No supported format specification or HTTP semantics rule reliably discriminates all supported payload formats from bytes alone.","interpretations":["Guess from names or media types.","Try every codec in order.","Recognize only unambiguous leading JSON and XML shapes."],"peer_behavior":"Maintained-peer automatic detection has not been assessed because explicit format selection is the normal API.","selected_behavior":"Detect only leading JSON objects or arrays and leading XML markup after whitespace, reporting every other payload as unsupported.","rationale":"Scalar JSON, YAML, TOML, and binary formats overlap and cannot be guessed reliably.","security_consequences":"The package does not execute multiple parsers speculatively on attacker-controlled input.","resource_consequences":"Detection performs bounded leading-byte inspection only.","compatibility_consequences":"Callers must explicitly select every format outside the narrow JSON and XML shapes.","wire_consequences":"SOAP is reported as XML and unknown or empty input remains unsupported.","executable_evidence":["TestDetectFormat","TestDetectFormatRejectsUnknownAndEmptyPayloads"],"fixture_evidence":[],"fuzz_evidence":[],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["DetectFormat"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"RFC 9110 defines media types but no content-sniffing algorithm for these formats.","reconsider_when":"A separate evidence-backed media-type boundary is introduced."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc9110.txt

</details>

**Authoritative reference:** [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html).

- **Status, owner, and classification:** `resolved`; maintainers; defensive
  application policy.
- **Source and issue:** No supported specification defines reliable automatic
  discrimination among JSON, XML, YAML, TOML, MessagePack, CBOR, and BSON.
  Several valid inputs are lexically ambiguous.
- **Interpretations and peer behavior:** Guess from extensions, MIME types,
  complete parsing, or leading bytes; try codecs in order; or require callers
  to select a format. Generic serializers commonly overstate detection.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  `DetectFormat` recognizes only a
  leading JSON object or array and a leading XML `<` after whitespace. It does
  not validate the payload, reports SOAP as XML, and never guesses scalar JSON,
  YAML, TOML, MessagePack, CBOR, or BSON. Unknown and empty input remain
  `ErrUnsupportedFormat`; explicit format selection is the normal API.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDetectFormat` and `TestDetectFormatRejectsUnknownAndEmptyPayloads` cover
  `DetectFormat`. No upstream standard exists; reconsider only for a separate
  evidence-backed media-type boundary.

## WIRE-DEC-005: JSON names, numbers, Unicode, and normalization

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-005","title":"JSON names, numbers, Unicode, and normalization","status":"resolved","owner":"wire maintainers","classification":"interoperability policy","decision_scope":"recommended","specification":"RFC 8259 JSON","version":"RFC 8259","source_authority":"rfc8259-source","section":"Sections 4, 6, and 8.1","requirement_strength":"SHOULD","issue":"RFC 8259 recommends unique names, permits number-range variation, and requires UTF-8 on open ecosystems while Go accepts duplicate names and invalid UTF-8.","interpretations":["Reject duplicate names unconditionally.","Preserve first or last duplicate values.","Expose compatibility defaults with explicit strictness."],"peer_behavior":"Go encoding/json replaces invalid UTF-8 and uses last-value duplicate handling, which deliberately differs from strict package options.","selected_behavior":"Require valid UTF-8, preserve Go numeric conversion, retain last-value duplicate compatibility by default, and provide recursive duplicate rejection explicitly.","rationale":"The default preserves established Go compatibility while exposing a bounded strict boundary for ambiguous objects.","security_consequences":"Strict mode prevents duplicate-member smuggling and invalid UTF-8 parser differentials.","resource_consequences":"Duplicate detection and normalization remain bounded by input limits.","compatibility_consequences":"Existing last-value behavior remains available while strict consumers can reject duplicates before destination mutation.","wire_consequences":"Normalize preserves number lexemes, orders keys, compacts whitespace, and strips only an initial UTF-8 BOM.","executable_evidence":["TestDuplicateNameValidationDefinesEveryJSONShape","TestJSONRejectsInvalidUTF8AcceptedByStandardLibrary","TestDecodeRejectsMalformedAndTrailingValues","TestNormalizeMakesVendorJSONCanonical"],"fixture_evidence":["jsonwire/testdata/valid.json","jsonwire/testdata/malformed.json"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":["specification/interoperability.tsv"],"public_apis":["jsonwire.Decode","jsonwire.Normalize"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"RFC 8259 errata are monitored; no stricter duplicate-name requirement is adopted.","reconsider_when":"A compatibility-governed profile changes the duplicate-name default or JSON source authority."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc8259.txt

</details>

**Authoritative reference:** [RFC 8259](https://www.rfc-editor.org/rfc/rfc8259.html).

- **Status, owner, and classification:** `resolved`; maintainers; RFC 8259
  interoperability policy with explicit compatibility mode.
- **Source and issue:** RFC 8259 [objects](https://www.rfc-editor.org/rfc/rfc8259.html#section-4)
  says names SHOULD be unique and permits implementations to differ for
  duplicates and number range; its encoding rules require UTF-8 on open
  ecosystems. Go's decoder accepts duplicate names and invalid UTF-8.
- **Interpretations and peer behavior:** Reject duplicates unconditionally,
  preserve first or last values, expose every pair, replace malformed Unicode,
  or follow Go defaults. Implementations differ on number precision and BOMs.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  Decode requires valid UTF-8 and one
  value. Duplicate-name rejection is explicit and recursive through
  `DisallowDuplicateNames`; compatibility default retains Go's last-value
  behavior. Typed decoding follows Go numeric conversion. `Normalize` alone
  strips an initial UTF-8 BOM, preserves number lexemes, compacts whitespace,
  orders object keys, and does not silently enable duplicate rejection.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDuplicateNameValidationDefinesEveryJSONShape`,
  `TestJSONRejectsInvalidUTF8AcceptedByStandardLibrary`,
  `TestDecodeRejectsMalformedAndTrailingValues`, and
  `TestNormalizeMakesVendorJSONCanonical` cover `Decode` and `Normalize`.
  Reconsider duplicate defaults only as a compatibility-governed change.

## WIRE-DEC-006: XML strictness, namespaces, entities, and charsets

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-006","title":"XML strictness, namespaces, entities, and charsets","status":"resolved","owner":"wire maintainers","classification":"implementation-defined behavior","decision_scope":"defensive","specification":"XML 1.0 Fifth Edition and Namespaces in XML 1.0 Third Edition","version":"XML 1.0 Fifth Edition","source_authority":"xml10-source","section":"XML sections 2 through 4 and Namespaces sections 2 through 6","requirement_strength":"not specified","issue":"Go exposes recovery and charset hooks while XML and Namespaces define syntax and expanded names without selecting this package's safe profile.","interpretations":["Recover malformed markup by default.","Compare lexical prefixes.","Use strict resolved names and explicit charset conversion."],"peer_behavior":"Maintained-peer recovery, DTD, and charset behavior has not been assessed as a package interoperability guarantee.","selected_behavior":"Use strict parsing, one root, resolved namespace names, finite depth, no declared entity expansion, and an explicit bounded charset allowlist.","rationale":"Resolved names and explicit decoding avoid prefix, entity, and charset ambiguity at service boundaries.","security_consequences":"DTD entity expansion and unbounded recovery are not enabled by default.","resource_consequences":"Token depth and input bytes are bounded before recursive work.","compatibility_consequences":"Non-strict recovery and custom charset readers remain explicit opt-ins.","wire_consequences":"Output follows encoding/xml traversal and does not claim XML canonicalization or lexical prefix preservation.","executable_evidence":["TestDecodeNamespaceAwareFixture","TestDecodeCanExplicitlyRecoverNonStrictXML","TestDecodeDoesNotExpandDeclaredEntities","TestCharsetReaderContract","TestEncodeIsDeterministicAndConfigurable"],"fixture_evidence":["xmlwire/testdata/shipment.xml","xmlwire/testdata/malformed.xml"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["xmlwire.Decode","xmlwire.Encode","xmlwire.Root"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"The dated XML and Namespaces editions are pinned and monitored independently.","reconsider_when":"The Go XML contract, required charset set, or supported XML edition changes."}
```

Authority URL: https://www.w3.org/TR/2008/REC-xml-20081126/

Additional authoritative source: `{"id":"xml-names-source","version":"Namespaces in XML 1.0 Third Edition","url":"https://www.w3.org/TR/2009/REC-xml-names-20091208/","specifications":["XML 1.0 Fifth Edition and Namespaces in XML 1.0 Third Edition"]}`

</details>

**Authoritative reference:** [XML 1.0 Fifth Edition](https://www.w3.org/TR/2008/REC-xml-20081126/).

- **Status, owner, and classification:** `resolved`; maintainers; XML 1.0
  interoperability and defensive profile.
- **Source and issue:** XML 1.0 Fifth Edition
  [defines document syntax](https://www.w3.org/TR/2008/REC-xml-20081126/) and
  Namespaces in XML 1.0 Third Edition
  [defines resolved names](https://www.w3.org/TR/2009/REC-xml-names-20091208/).
  Go exposes strict recovery and caller-provided charset conversion without
  canonical XML.
- **Interpretations and peer behavior:** Recover malformed markup by default,
  compare lexical prefixes, expand declarations, guess encodings, or require
  strict resolved names and declared conversion. XML stacks vary materially on
  DTD and entity behavior.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  Strict parsing, one root, resolved
  namespace URL plus local-name comparison, finite token depth, and no declared
  entity expansion are defaults. Non-strict recovery is opt-in. Built-in
  charset conversion is limited to UTF-8, ASCII, ISO-8859-1, and Windows-1252;
  other declared encodings require an explicit reader. Output follows
  `encoding/xml` struct traversal and is deterministic for that value, but is
  not XML canonicalization and does not preserve lexical prefixes.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDecodeNamespaceAwareFixture`,
  `TestDecodeCanExplicitlyRecoverNonStrictXML`,
  `TestDecodeDoesNotExpandDeclaredEntities`, `TestCharsetReaderContract`, and
  `TestEncodeIsDeterministicAndConfigurable` cover `xmlwire`. Reconsider when
  Go's XML contract or a required vendor charset changes.

## WIRE-DEC-007: SOAP versions, envelope profile, and faults

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-007","title":"SOAP versions, envelope profile, and faults","status":"resolved","owner":"wire maintainers","classification":"implementation-defined behavior","decision_scope":"transport-specific","specification":"SOAP 1.1 and SOAP 1.2 Part 1 Second Edition","version":"SOAP 1.2 Part 1 Second Edition","source_authority":"soap12-source","section":"SOAP 1.1 section 4 and SOAP 1.2 sections 2 through 5","requirement_strength":"not specified","issue":"SOAP 1.1 and 1.2 define different envelope and fault structures while processing, headers, and HTTP policy extend beyond a transport-neutral codec.","interpretations":["Coerce both versions into one loose XML shape.","Bind HTTP and WS-* policy.","Expose a strict envelope profile with raw escape hatches."],"peer_behavior":"Maintained SOAP-stack behavior has not been assessed; local fixtures cover the selected transport-neutral profile.","selected_behavior":"Select the version by namespace, require the defined envelope structure, preserve raw sections, and expose version-specific typed faults without owning HTTP or WS-* policy.","rationale":"A narrow envelope codec avoids silently implementing an incomplete SOAP processing model.","security_consequences":"Unexpected envelope structure and malformed raw fragments fail before application decoding.","resource_consequences":"Envelope parsing and raw section copies remain bounded by package limits.","compatibility_consequences":"SOAP 1.1 and 1.2 faults retain their distinct public fields and namespace rules.","wire_consequences":"Typed bodies require exactly one child and valid faults return both the envelope and FaultError.","executable_evidence":["TestParseSOAP11EnvelopePreservesRawSections","TestParseSOAP12FaultReturnsTypedErrorAndEnvelope","TestParseRejectsInvalidEnvelopeStructure","TestDecodeBodyRetainsInheritedNamespaces","TestMarshalFaultRoundTripsBothVersions"],"fixture_evidence":["soap/testdata/soap11-response.xml","soap/testdata/soap11-fault.xml","soap/testdata/soap12-fault.xml"],"fuzz_evidence":["FuzzParse"],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["soap.Parse","soap.Encode","soap.FaultError"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"The dated SOAP 1.1 and SOAP 1.2 publications are pinned and monitored together.","reconsider_when":"A separately specified processing or transport profile is added."}
```

Authority URL: https://www.w3.org/TR/2007/REC-soap12-part1-20070427/

Additional authoritative source: `{"id":"soap11-source","version":"SOAP 1.1","url":"https://www.w3.org/TR/2000/NOTE-SOAP-20000508/","specifications":["SOAP 1.1 and SOAP 1.2 Part 1 Second Edition"]}`

</details>

**Authoritative reference:** [SOAP 1.2 Part 1 Second Edition](https://www.w3.org/TR/2007/REC-soap12-part1-20070427/).

- **Status, owner, and classification:** `resolved`; maintainers; SOAP 1.1/1.2
  transport-neutral profile.
- **Source and issue:** SOAP 1.1
  [section 4](https://www.w3.org/TR/2000/NOTE-SOAP-20000508/#_Toc478383494)
  and SOAP 1.2 Part 1 Second Edition
  [section 5](https://www.w3.org/TR/2007/REC-soap12-part1-20070427/#soapfault)
  define different envelope and fault structures, while actor/role, headers,
  body content, HTTP action, and processing models can involve transport and
  application policy.
- **Interpretations and peer behavior:** Coerce both versions into one loose
  XML shape, preserve arbitrary envelope children, bind HTTP behavior, or
  expose a strict request/response profile with raw escape hatches.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  The namespace selects exactly SOAP
  1.1 or 1.2. Header is optional and precedes one required Body; unexpected
  envelope children or text fail. Typed body decoding requires exactly one
  child; raw sections remain copied bytes with inherited namespaces preserved
  for decoding. A valid fault returns both the envelope and `*FaultError`, with
  version-specific code, reasons, subcodes, actor/node/role, and Detail. HTTP,
  SOAPAction, header processing, WSDL, WS-* and application mapping are outside
  this package.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestParseSOAP11EnvelopePreservesRawSections`,
  `TestParseSOAP12FaultReturnsTypedErrorAndEnvelope`,
  `TestParseRejectsInvalidEnvelopeStructure`,
  `TestDecodeBodyRetainsInheritedNamespaces`, and
  `TestMarshalFaultRoundTripsBothVersions` cover `soap`. Reconsider for a new
  separately specified processing profile, not by weakening this one.

## WIRE-DEC-008: YAML schema, aliases, merges, duplicates, and streams

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-008","title":"YAML schema, aliases, merges, duplicates, and streams","status":"resolved","owner":"wire maintainers","classification":"ambiguity","decision_scope":"defensive","specification":"YAML 1.2.2","version":"YAML 1.2.2","source_authority":"yaml122-source","section":"Chapters 3 and 10","requirement_strength":"not specified","issue":"YAML graphs, recommended schemas, aliases, tags, merge behavior, and multi-document streams do not map uniformly to native Go values.","interpretations":["Use permissive YAML 1.1 typing.","Reject every advanced feature.","Adopt pinned v4 defaults with explicit defensive options."],"peer_behavior":"go.yaml.in/yaml/v4 is the maintained codec; the wrapper repairs a known invalid block-scalar indentation output.","selected_behavior":"Use v4 defaults, one document and unique keys by default, sorted maps, finite protections, and explicit options for streams, aliases, and merges.","rationale":"The package preserves YAML's supported model without claiming a JSON-compatible subset or unbounded graph support.","security_consequences":"Duplicate keys, hostile aliases, and excessive depth are rejected by default protections.","resource_consequences":"Alias, depth, input, and output limits bound processing.","compatibility_consequences":"Tags, implicit types, non-string keys, aliases, and merges remain documented codec-governed behavior.","wire_consequences":"Multiple documents require an explicit slice target and opt-in; emitted mapping keys are sorted.","executable_evidence":["TestDecodeRejectsMalformedDuplicateAndMultipleDocuments","TestDecodeDefinesAliasAnchorAndMergeBehavior","TestDecodeDefinesTagsImplicitTypesAndNonJSONKeys","TestDecodeClassifiesBuiltInResourceProtectionAsSizeLimit","TestYAMLRepairsDependencyBlockIndentDifferential"],"fixture_evidence":["yamlwire/testdata/manifest.yaml"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":["specification/interoperability.tsv"],"public_apis":["yamlwire.Decode","yamlwire.Encode"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"The YAML edition and maintained codec tags are monitored separately.","reconsider_when":"The YAML edition, default schema, or codec behavior changes."}
```

Authority URL: https://raw.githubusercontent.com/yaml/yaml-spec/80c6bde43887cf83424defb5c45c2485464659b9/spec/1.2.2/spec.md

</details>

**Authoritative reference:** [YAML 1.2.2](https://yaml.org/spec/1.2.2/).

- **Status, owner, and classification:** `resolved`; maintainers; YAML 1.2.2
  interoperability and defensive policy.
- **Source and issue:** YAML 1.2.2
  [chapters 3 and 10](https://yaml.org/spec/1.2.2/) define representation,
  serialization, presentation, and recommended schema models with tags,
  anchors, aliases, and multi-document streams. Native Go mapping cannot
  preserve every YAML graph or key shape uniformly; merge keys are a codec
  interoperability feature rather than YAML 1.2.2 core syntax.
- **Interpretations and peer behavior:** Use permissive YAML 1.1 typing, reject
  every advanced feature, expose dependency defaults, or select explicit safe
  v4 behavior and options. YAML codecs differ on duplicate keys and alias
  limits.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  Use `go.yaml.in/yaml/v4` v4 defaults,
  one document and unique keys by default, sorted mapping output, and codec
  depth/alias protections. Multiple documents require a slice and opt-in;
  aliases and merge keys may be independently rejected, and caller limits may
  be stricter. Tags, core-schema implicit values, and typed non-string map keys
  remain dependency-governed and documented; no JSON-compatible-subset claim
  is made.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDecodeRejectsMalformedDuplicateAndMultipleDocuments`,
  `TestDecodeDefinesAliasAnchorAndMergeBehavior`,
  `TestDecodeDefinesTagsImplicitTypesAndNonJSONKeys`,
  `TestDecodeClassifiesBuiltInResourceProtectionAsSizeLimit`, and
  `TestYAMLRepairsDependencyBlockIndentDifferential` cover `yamlwire`.
  Reconsider every YAML dependency or default-schema change.

## WIRE-DEC-009: TOML document model and conversion

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-009","title":"TOML document model and conversion","status":"resolved","owner":"wire maintainers","classification":"implementation-defined behavior","decision_scope":"application-policy","specification":"TOML 1.1.0","version":"TOML 1.1.0","source_authority":"toml110-source","section":"Complete specification","requirement_strength":"not specified","issue":"TOML defines one document and native value types but not Go target conversion, unknown-field policy, or source-format preservation.","interpretations":["Treat TOML as generic maps.","Coerce out-of-range values.","Preserve native types and expose strict target validation."],"peer_behavior":"Maintained BurntSushi/toml behavior is delegated but no separate peer differential is currently assessed.","selected_behavior":"Parse one complete document, preserve native date, time, and numeric types, reject lossy target conversion, and make unknown-field rejection explicit.","rationale":"Typed conversion must not silently lose TOML value identity or numeric range.","security_consequences":"Conflicting keys and incompatible target assignments fail instead of being normalized silently.","resource_consequences":"Input and encoded output remain bounded by package limits.","compatibility_consequences":"The valid empty document remains accepted and strict unknown-field checking stays opt-in.","wire_consequences":"Encoding sorts map keys but does not preserve comments or original formatting.","executable_evidence":["TestDecodeFixturePreservesDatetimeAndNumericTypes","TestDecodeRejectsMalformedDuplicateAndTrailingData","TestDecodeRejectsUnknownFieldsAndNumericLossWhenRequested","TestEncodeIsDeterministicAndPreservesNativeTypes"],"fixture_evidence":["tomlwire/testdata/service.toml"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["tomlwire.Decode","tomlwire.Encode"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"TOML 1.1.0 and upstream tags are pinned and monitored.","reconsider_when":"The TOML edition or maintained codec changes."}
```

Authority URL: https://raw.githubusercontent.com/toml-lang/toml/bcbbd1c1f03473ffe97b8bf26a0fc945efe2b4a1/toml.md

</details>

**Authoritative reference:** [TOML 1.1.0](https://toml.io/en/v1.1.0).

- **Status, owner, and classification:** `resolved`; maintainers; TOML 1.1.0
  interoperability policy.
- **Source and issue:** TOML 1.1.0
  [defines](https://toml.io/en/v1.1.0) one UTF-8 configuration document with
  unique keys, tables, date/time values, and bounded numeric grammar, while Go
  target conversion and unknown-field handling remain application choices.
- **Interpretations and peer behavior:** Treat TOML as generic maps, coerce
  out-of-range values, ignore unknown keys, or preserve native TOML types and
  expose strict target validation.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  `BurntSushi/toml` v1.6.0 parses one
  complete document, including the valid empty document, and rejects duplicate
  or conflicting keys. Native date/time and numeric types are preserved;
  narrowing or incompatible target conversion is validation failure. Unknown
  fields are optional strictness. Encoding sorts map keys and supports explicit
  whitespace-only indentation, but does not promise source formatting or
  comments round-trip.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDecodeFixturePreservesDatetimeAndNumericTypes`,
  `TestDecodeRejectsMalformedDuplicateAndTrailingData`,
  `TestDecodeRejectsUnknownFieldsAndNumericLossWhenRequested`, and
  `TestEncodeIsDeterministicAndPreservesNativeTypes` cover `tomlwire`.
  Reconsider on TOML edition or codec changes.

## WIRE-DEC-010: MessagePack maps, extensions, numeric fit, and structure

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-010","title":"MessagePack maps, extensions, numeric fit, and structure","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"defensive","specification":"MessagePack format at 8aa09e2a6a91","version":"MessagePack commit 8aa09e2a6a91","source_authority":"msgpack-source","section":"Complete format specification","requirement_strength":"not specified","issue":"MessagePack defines encoded values but not duplicate-key handling, Go numeric narrowing, application extension ownership, or resource limits.","interpretations":["Accept last-key-wins maps and coercion.","Trust declared lengths.","Validate duplicates, numeric fit, structure, and extensions explicitly."],"peer_behavior":"github.com/vmihailenco/msgpack/v5 accepts duplicate map keys that the wrapper rejects by default.","selected_behavior":"Require one object, reject recursive duplicates by default, validate typed numeric fit and structural limits, and reject unknown extensions.","rationale":"Ambiguous maps, forged lengths, and lossy conversion are unsafe defaults for untrusted data.","security_consequences":"Duplicate keys, unknown extensions, and hostile lengths fail before ambiguous assignment or allocation.","resource_consequences":"Finite nesting and collection limits preflight declared sizes.","compatibility_consequences":"Last-key-wins compatibility and alternate struct or numeric modes remain explicit options.","wire_consequences":"Encoders sort supported map keys and preserve registered timestamp behavior.","executable_evidence":["TestDecodeRejectsMalformedTrailingAndUnknownExtensions","TestDecodeRejectsDuplicateMapKeysByDefault","TestDecodeValidatesNestedNumericAssignments","TestDecodeEnforcesDefaultStructuralLimits","TestEncodeIsDeterministicAndConfigurable"],"fixture_evidence":["msgpackwire/testdata/shipment.msgpack.hex"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":["specification/interoperability.tsv"],"public_apis":["msgpackwire.Decode","msgpackwire.Encode"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"The exact format commit and maintained peer tags are monitored.","reconsider_when":"Extension ownership or the maintained codec contract changes."}
```

Authority URL: https://raw.githubusercontent.com/msgpack/msgpack/8aa09e2a6a9180a49fc62ecfefe149f063cc5e4b/spec.md

</details>

**Authoritative reference:** [MessagePack format specification](https://github.com/msgpack/msgpack/blob/8aa09e2a6a9180a49fc62ecfefe149f063cc5e4b/spec.md).

- **Status, owner, and classification:** `resolved`; maintainers; MessagePack
  interoperability and defensive policy.
- **Source and issue:** The pinned MessagePack
  [format specification](https://github.com/msgpack/msgpack/blob/8aa09e2a6a9180a49fc62ecfefe149f063cc5e4b/spec.md)
  defines compact values, arbitrary map keys, extension identifiers, and
  multiple integer widths, but not duplicate-key resolution, Go assignment
  narrowing, application extension registries, or resource limits.
- **Interpretations and peer behavior:** Accept last-key-wins maps, coerce
  numbers, decode all keys into interfaces, trust declared lengths, or validate
  structure and target fit before assignment.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  Exactly one object is required.
  Recursive duplicates fail by default with opt-in compatibility; untyped maps
  require string keys while typed comparable keys remain available. Unknown
  extensions are unsupported, registered timestamp behavior is retained, and
  numeric overflow or precision loss into typed destinations fails before the
  main assignment. Explicit finite nesting and collection limits preflight
  forged lengths. Encoders sort map keys; compact integer, float, and array-
  struct modes are explicit.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDecodeRejectsMalformedTrailingAndUnknownExtensions`,
  `TestDecodeRejectsDuplicateMapKeysByDefault`,
  `TestDecodeValidatesNestedNumericAssignments`,
  `TestDecodeEnforcesDefaultStructuralLimits`, and
  `TestEncodeIsDeterministicAndConfigurable` cover `msgpackwire`. Reconsider
  when extension ownership or a codec with native safe limits is adopted.

## WIRE-DEC-011: CBOR deterministic profiles and accepted data items

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-011","title":"CBOR deterministic profiles and accepted data items","status":"resolved","owner":"wire maintainers","classification":"optional behavior","decision_scope":"application-policy","specification":"RFC 7049 and RFC 8949 CBOR","version":"RFC 8949","source_authority":"rfc8949-source","section":"RFC 7049 section 3.9, RFC 8949 section 4.2, and CTAP 2.2 section 8","requirement_strength":"not specified","issue":"RFC 7049, RFC 8949, and CTAP2 define different deterministic profiles while tags, indefinite lengths, and accepted data items remain independently configurable.","interpretations":["Call every stable encoding canonical.","Accept every valid item by default.","Expose exact profiles and independent decode policy."],"peer_behavior":"github.com/fxamacker/cbor/v2 accepts duplicate keys under its generic default while the wrapper rejects them.","selected_behavior":"Name the three exact deterministic profiles, reject duplicate keys, keep tags and indefinite lengths opt-in, and retain finite limits.","rationale":"Profile names must preserve their source semantics instead of collapsing incompatible canonicalization rules.","security_consequences":"Duplicate-key ambiguity and unbounded structures are rejected independently of profile selection.","resource_consequences":"Input, nesting, collection, and output limits remain finite.","compatibility_consequences":"Legacy Canonical remains the encode default while other profiles require explicit selection.","wire_consequences":"Decode does not claim that incoming bytes already satisfy any deterministic encode profile.","executable_evidence":["TestEncodeUsesExplicitDeterministicProfiles","TestDecodeDefinesTagsAndIndefiniteLengthBehavior","TestDecodePreservesSimpleValuesAndBignums","TestDecodeEnforcesUnknownFieldsNumericAndResourceLimits"],"fixture_evidence":["cborwire/testdata/shipment.cbor.hex"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":["specification/interoperability.tsv"],"public_apis":["cborwire.Decode","cborwire.Encode","cborwire.Profile"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"RFC errata, CTAP publications, and the maintained codec remain separately monitored.","reconsider_when":"Profile naming, minimum versions, or accepted data-item policy changes."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc8949.txt

Additional authoritative source: `{"id":"rfc7049-source","version":"RFC 7049","url":"https://www.rfc-editor.org/rfc/rfc7049.txt","specifications":["RFC 7049 and RFC 8949 CBOR"]}`

Additional authoritative source: `{"id":"ctap22-source","version":"CTAP 2.2 Proposed Standard 2025-07-14","url":"https://fidoalliance.org/specs/fido-v2.2-ps-20250714/fido-client-to-authenticator-protocol-v2.2-ps-20250714.pdf","specifications":["CTAP 2.2 deterministic CBOR profile"]}`

</details>

**Authoritative reference:** [RFC 8949](https://www.rfc-editor.org/rfc/rfc8949.html).

- **Status, owner, and classification:** `resolved`; maintainers; RFC 7049,
  RFC 8949, and CTAP2 profile policy.
- **Source and issue:** RFC 7049
  [section 3.9](https://www.rfc-editor.org/rfc/rfc7049.html#section-3.9), RFC
  8949 [section 4.2](https://www.rfc-editor.org/rfc/rfc8949.html#section-4.2),
  and CTAP 2.2 [section 8](https://fidoalliance.org/specs/fido-v2.2-ps-20250714/fido-client-to-authenticator-protocol-v2.2-ps-20250714.html#sctn-encoded-message)
  define different deterministic profiles. CBOR tags, indefinite lengths,
  duplicate keys, simple values, bignums, and preferred serialization are
  separately configurable.
- **Interpretations and peer behavior:** Call every stable encoding canonical,
  accept all valid CBOR by default, silently normalize profiles, or expose the
  exact three profiles and independent decode policy.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  `Canonical` means RFC 7049 section
  3.9, `CoreDeterministic` means RFC 8949 core deterministic encoding, and
  `CTAP2Deterministic` means the pinned CTAP2 profile. Decode always rejects
  duplicate keys; tags and indefinite lengths are independently opt-in, while
  resource bounds remain finite. Encode defaults to legacy `Canonical`; tags
  require explicit permission and time tags require tags. Valid simple values
  and tagged bignums retain codec-defined Go representations. No decoder claim
  is made that incoming bytes already satisfy an encode profile.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestEncodeUsesExplicitDeterministicProfiles`,
  `TestDecodeDefinesTagsAndIndefiniteLengthBehavior`,
  `TestDecodePreservesSimpleValuesAndBignums`, and
  `TestDecodeEnforcesUnknownFieldsNumericAndResourceLimits` cover `cborwire`.
  Reconsider if profile naming or minimum-version defaults change.

## WIRE-DEC-012: BSON document identity, order, duplicates, and conversion

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-012","title":"BSON document identity, order, duplicates, and conversion","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"defensive","specification":"BSON 1.1","version":"BSON 1.1","source_authority":"bson11-source","section":"Complete grammar and type specification","requirement_strength":"not specified","issue":"BSON defines ordered typed documents but does not select duplicate-name handling, Go map determinism, or lossy target conversion policy.","interpretations":["Accept scalars and trailing bytes.","Use last-value maps.","Require one document and explicit ordered or compatibility representations."],"peer_behavior":"The maintained MongoDB Go driver preserves duplicate elements in ordered documents while the wrapper rejects them by default.","selected_behavior":"Require one complete top-level document, reject recursive duplicates by default, preserve official driver types, and expose lossy conversions only as options.","rationale":"Document identity and order must remain explicit and ambiguous duplicate data must not silently collapse.","security_consequences":"Declared length, terminator, trailing bytes, and duplicate names are validated before assignment.","resource_consequences":"Input and output remain bounded and forged lengths are rejected before large allocation.","compatibility_consequences":"Ordered D and raw forms remain stable while unordered M output is explicitly non-deterministic.","wire_consequences":"ObjectID string conversion, truncating doubles, integer minimization, and JSON tags are opt-in.","executable_evidence":["TestDecodeRejectsMalformedTrailingDuplicateAndScalarData","TestDecodeRejectsNestedDuplicateKeys","TestDecodeProvidesExplicitInteroperabilityOptions","TestRoundTripPreservesDecimalBinarySubtypeAndRegex","TestEncodeOrderedDocumentsAreDeterministic"],"fixture_evidence":["bsonwire/testdata/event.bson.hex"],"fuzz_evidence":["FuzzDecode"],"interoperability_evidence":[],"differential_evidence":["specification/interoperability.tsv"],"public_apis":["bsonwire.Decode","bsonwire.Encode","bsonwire.D","bsonwire.M"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"The BSON 1.1 publication and maintained driver are monitored.","reconsider_when":"The BSON edition or official driver type model changes."}
```

Authority URL: https://bsonspec.org/spec.html

</details>

**Authoritative reference:** [BSON Specification Version 1.1](https://bsonspec.org/spec.html).

- **Status, owner, and classification:** `resolved`; maintainers; BSON 1.1 and
  official-driver interoperability policy.
- **Source and issue:** BSON
  [Specification Version 1.1](https://bsonspec.org/spec.html) defines an
  ordered document with length prefixes and typed elements, while generic Go
  maps lose ordering and the specification does not select duplicate-name
  resolution or lossy target conversion.
- **Interpretations and peer behavior:** Accept scalars, ignore trailing bytes,
  accept duplicate names, claim map determinism, create local BSON types, or
  retain the official driver model with explicit options.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  APIs require exactly one complete
  top-level document whose declared length and terminator validate. Recursive
  duplicate names fail by default with explicit compatibility opt-in.
  Official driver types are re-exported rather than copied. Struct, `D`, and
  raw order are stable; `M` map order is explicitly not deterministic.
  ObjectID-as-string and truncating-double conversion, integer-width
  minimization, and JSON struct tags are explicit options rather than defaults.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestDecodeRejectsMalformedTrailingDuplicateAndScalarData`,
  `TestDecodeRejectsNestedDuplicateKeys`,
  `TestDecodeProvidesExplicitInteroperabilityOptions`,
  `TestRoundTripPreservesDecimalBinarySubtypeAndRegex`, and
  `TestEncodeOrderedDocumentsAreDeterministic` cover `bsonwire`. Reconsider on
  BSON edition or official driver type changes.

## WIRE-DEC-013: Deterministic output is not universal canonicalization

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-013","title":"Deterministic output is not universal canonicalization","status":"resolved","owner":"wire maintainers","classification":"implementation-defined behavior","decision_scope":"application-policy","specification":"RFC 7049 and RFC 8949 CBOR","version":"RFC 8949","source_authority":"rfc8949-source","section":"Section 4.2 deterministic encoding","requirement_strength":"not specified","issue":"Some formats define deterministic profiles while others permit many equivalent serializations, so stable bytes cannot be called universally canonical.","interpretations":["Claim every encoder is canonical.","Promise semantic round-trip only.","Describe determinism per format and value shape."],"peer_behavior":"Maintained-peer byte agreement has not been assessed as a cross-format guarantee.","selected_behavior":"Claim sorted or stable output only for documented value shapes and name canonicalization only where an exact source profile exists.","rationale":"Signing and hashing consumers must not infer a standards-level canonical form from incidental stable bytes.","security_consequences":"The documentation avoids unsafe cross-format signing assumptions.","resource_consequences":"Deterministic modes remain subject to the same bounded output policy.","compatibility_consequences":"Codec upgrades and formatting options remain compatibility-relevant even when semantic values are unchanged.","wire_consequences":"XML is not C14N, BSON maps are unordered, and CBOR names the selected deterministic profile exactly.","executable_evidence":["TestEncodeIsDeterministicAndConfigurable","TestEncodeIsDeterministicAndPreservesNativeTypes","TestEncodeUsesExplicitDeterministicProfiles","TestEncodeOrderedDocumentsAreDeterministic"],"fixture_evidence":[],"fuzz_evidence":["FuzzRoundTrip"],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["All Encode and EncodeWriter functions"],"documentation":["docs/specification-decisions.md","docs/formats.md"],"upstream_status":"Only CBOR sources define the named deterministic profiles used by this package.","reconsider_when":"A signing use case or codec upgrade proposes a stronger byte-level guarantee."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc8949.txt

</details>

**Authoritative reference:** [RFC 8949](https://www.rfc-editor.org/rfc/rfc8949.html).

- **Status, owner, and classification:** `resolved`; maintainers; wire-format
  compatibility policy.
- **Source and issue:** Some formats define deterministic or canonical
  encodings, while others permit many semantically equivalent serializations.
  Stable current bytes can be mistaken for a standards-level canonical form.
- **Interpretations and peer behavior:** Claim every encoder is canonical,
  promise only semantic round-trip, or describe determinism per format and
  value shape.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  JSON, YAML, TOML, and MessagePack
  sort supported map keys; XML follows stable struct traversal but is not C14N;
  SOAP uses fixed package-owned envelope emission; CBOR names its exact
  profile; BSON guarantees order only for ordered shapes. Formatting options,
  codec upgrades, semantically unordered maps, and dependency representations
  remain compatibility-relevant. Stable bytes are claimed only where the
  format matrix says so.
- **Evidence, public surface, upstream, and reconsideration:** Per-format
  `TestEncodeIsDeterministicAndConfigurable`,
  `TestEncodeIsDeterministicAndPreservesNativeTypes`,
  `TestEncodeUsesExplicitDeterministicProfiles`, and
  `TestEncodeOrderedDocumentsAreDeterministic` cover encoder options. Reconsider
  every codec upgrade or proposed signing/hash use.

## WIRE-DEC-014: Error classification, causes, and disclosure

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-014","title":"Error classification, causes, and disclosure","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"defensive","specification":"Go 1.26.6 errors package contract","version":"Go 1.26.6","source_authority":"go-errors-source","section":"src/errors and errors wrapping contract","requirement_strength":"not specified","issue":"Codec errors mix syntax, conversion, limits, targets, protocol faults, and raw diagnostics without a shared stable package taxonomy.","interpretations":["Return codec errors directly.","Expose only sentinels.","Wrap bounded causes in a stable cross-format taxonomy."],"peer_behavior":"Maintained-peer diagnostic disclosure has not been assessed as a portable contract.","selected_behavior":"Classify stable error kinds while preserving errors.Is and errors.As causes without echoing complete payloads.","rationale":"Callers need actionable classification without coupling to dependency text or leaking attacker-controlled data.","security_consequences":"Complete payloads and tested sensitive values are not included in errors.","resource_consequences":"Error construction does not add unbounded copies of input data.","compatibility_consequences":"Stable sentinels and typed errors remain usable across dependency error-shape changes.","wire_consequences":"Valid SOAP faults remain protocol outcomes rather than malformed input.","executable_evidence":["TestErrorKindsMatchTheirSentinels","TestErrorSupportsClassificationAndWrapping","TestDecodeErrorsDoNotEchoSensitiveValues"],"fixture_evidence":[],"fuzz_evidence":[],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["Error","ErrorKind","FaultError","All exported error sentinels"],"documentation":["docs/specification-decisions.md","docs/api.md"],"upstream_status":"The Go 1.26.6 errors contract and release stream are pinned and monitored.","reconsider_when":"A dependency error shape or disclosure boundary changes."}
```

Authority URL: https://api.github.com/repos/golang/go/contents/src/errors?ref=go1.26.6

</details>

**Authoritative reference:** [Go errors package](https://pkg.go.dev/errors).

- **Status, owner, and classification:** `resolved`; maintainers; defensive
  API and privacy policy.
- **Source and issue:** Codec errors mix syntax, type conversion, unsupported
  features, limits, destination failures, and valid SOAP faults. Raw errors can
  expose input fragments, while flattening loses actionable causes.
- **Interpretations and peer behavior:** Return codec errors directly, expose
  only sentinels, redact every diagnostic, or wrap bounded causes in a stable
  cross-format taxonomy.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  `wire.Error` classifies parse,
  validation, target, unsupported, envelope, SOAP fault, size, encode, and
  write outcomes with format and operation. `errors.Is` and `errors.As` retain
  stable classification and useful causes. The package does not echo complete
  payloads or the tested sensitive values; callers must still treat field names
  and small offending lexemes as potentially sensitive. Valid SOAP faults are
  protocol outcomes, not malformed input.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestErrorKindsMatchTheirSentinels`,
  `TestErrorSupportsClassificationAndWrapping`,
  `TestDecodeErrorsDoNotEchoSensitiveValues`, and SOAP fault tests cover
  `Error`, sentinels, and `FaultError`. Reconsider whenever a dependency error
  shape or disclosure boundary changes.

## WIRE-DEC-015: Encode graph validation and dependency differentials

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-015","title":"Encode graph validation and dependency differentials","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"defensive","specification":"Go 1.26.6 language specification","version":"Go 1.26.6","source_authority":"go-language-source","section":"Types, values, and pointer semantics","requirement_strength":"not specified","issue":"Recursive Go values can cycle or exceed stack-safe depth while maintained codecs may accept behavior that conflicts with package guarantees.","interpretations":["Trust codecs and recover panics.","Reject every shared reference.","Preflight active recursion paths and retain classified differentials."],"peer_behavior":"Maintained codec differences for invalid JSON UTF-8, duplicate maps, and YAML emission are recorded as executable evidence.","selected_behavior":"Reject cycles on the active path and traversal beyond 1000 levels while allowing reused acyclic references and retaining classified codec differentials.","rationale":"Package safety guarantees must be established before recursive dependency work without rejecting valid shared acyclic values.","security_consequences":"Cyclic or excessively deep values fail before stack exhaustion or unbounded recursive encoding.","resource_consequences":"Graph traversal has an explicit finite depth and active-path ownership model.","compatibility_consequences":"Known dependency differences remain deliberate package policy rather than hidden patches.","wire_consequences":"Only validated acyclic values reach the selected format encoder.","executable_evidence":["TestAllEncodePathsRejectCyclicValues","TestValidateAcceptsAcyclicValues","TestValidateRejectsCyclesAndDepth","TestJSONRejectsInvalidUTF8AcceptedByStandardLibrary","TestMessagePackRejectsDuplicateAcceptedByDependency","TestCBORRejectsDuplicateAcceptedByDependencyDefault","TestBSONRejectsDuplicatePreservedByDependency","TestYAMLRepairsDependencyBlockIndentDifferential"],"fixture_evidence":[],"fuzz_evidence":[],"interoperability_evidence":[],"differential_evidence":["specification/interoperability.tsv"],"public_apis":["All typed encode APIs"],"documentation":["docs/specification-decisions.md","docs/security.md"],"upstream_status":"The Go language edition and maintained codec versions are pinned and monitored.","reconsider_when":"A codec upgrade or new recursive value model changes graph behavior."}
```

Authority URL: https://raw.githubusercontent.com/golang/go/go1.26.6/doc/go_spec.html

</details>

**Authoritative reference:** [Go language specification](https://go.dev/ref/spec).

- **Status, owner, and classification:** `resolved`; maintainers; defensive
  Go-value and dependency-boundary policy.
- **Source and issue:** Recursive Go values can cycle or exceed stack-safe
  depth even where the wire format itself has no references. Mature codecs can
  also accept or emit behavior that conflicts with this package's guarantees.
- **Interpretations and peer behavior:** Trust codecs, recover panics, reject
  every shared reference, or preflight only active recursion paths and retain
  differential tests for known policy gaps.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  Every typed encoder rejects cycles
  on the active path and traversal beyond 1,000 levels before recursive codec
  work. Reused acyclic references remain valid. Package policy deliberately
  differs from dependencies for invalid JSON UTF-8, MessagePack, CBOR and BSON
  duplicate handling, and YAML block-scalar emission. Those differences are
  executable compatibility evidence, not hidden patches.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestAllEncodePathsRejectCyclicValues`,
  `TestValidateAcceptsAcyclicValues`,
  `TestValidateRejectsCyclesAndDepth`, and all tests in
  `dependency_differential_test.go` cover shared validation and codec seams.
  Reconsider on any dependency upgrade or new recursive value support.

## WIRE-DEC-016: Transport, schema, and application boundaries

<details>
<summary>Machine-auditable bindings</summary>

```json
{"id":"WIRE-DEC-016","title":"Transport, schema, and application boundaries","status":"resolved","owner":"wire maintainers","classification":"omission","decision_scope":"application-policy","specification":"RFC 9110 HTTP Semantics","version":"RFC 9110","source_authority":"rfc9110-source","section":"Sections 6 through 8","requirement_strength":"not specified","issue":"Wire formats are often coupled to HTTP, media types, schemas, signing, persistence, and application mapping that no single format specification owns uniformly.","interpretations":["Grow one serialization framework owning surrounding policy.","Infer transport behavior.","Remain a narrow composable format boundary."],"peer_behavior":"Maintained-peer framework scope has not been assessed because the decision defines this package's ownership boundary.","selected_behavior":"Own bounded byte and reader or writer format handling only, leaving transport, schema, signing, compression, persistence, and application policy to composed layers.","rationale":"A narrow boundary keeps unrelated standards and application lifecycle policy out of a codec package.","security_consequences":"Callers must validate media, schema, compression, signature, and trust policy at their owning boundaries.","resource_consequences":"Only codec-owned input, output, and structure budgets are enforced here.","compatibility_consequences":"New surrounding capabilities require separately specified additive APIs and decisions.","wire_consequences":"The package does not select HTTP content types, status codes, SOAPAction, JSON-RPC, JSON:API, WSDL, XSD, or registries.","executable_evidence":["TestSharedDocumentationConventions","TestSharedRepositoryContract"],"fixture_evidence":[],"fuzz_evidence":[],"interoperability_evidence":[],"differential_evidence":[],"public_apis":["Root wire package and all format packages"],"documentation":["docs/specification-decisions.md","docs/architecture.md"],"upstream_status":"RFC 9110 is monitored only for the HTTP boundary explicitly excluded from package ownership.","reconsider_when":"A separately specified transport, schema, or application API is proposed."}
```

Authority URL: https://www.rfc-editor.org/rfc/rfc9110.txt

</details>

**Authoritative reference:** [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html).

- **Status, owner, and classification:** `resolved`; maintainers; scope and
  ownership policy.
- **Source and issue:** Wire formats are commonly coupled to media types, HTTP,
  RPC, schemas, WSDL, vendor mapping, persistence, signing, or content
  negotiation, none of which is defined uniformly by the format grammars.
- **Interpretations and peer behavior:** Grow one serialization framework that
  owns all surrounding policy, infer transport behavior, or remain a narrow
  format boundary composed by higher layers.
- **Selected behavior and consequences:**
  Security, resource, compatibility, and wire consequences are included in
  the selected behavior below.
  `wire` owns bounded byte and
  reader/writer format handling only. It does not select HTTP content types or
  status codes, SOAPAction, JSON-RPC/JSON:API behavior, WSDL/XSD validation,
  schema registries, compression, signatures, persistence, or business
  normalization. Callers must select and validate those policies before or
  after this boundary. No known material ambiguity inside the current public
  surface remains unresolved.
- **Evidence, public surface, upstream, and reconsideration:**
  `TestSharedDocumentationConventions`, the format matrix, architecture docs,
  and package dependency graph cover the boundary. No single upstream source
  owns it; reconsider only through a separately specified additive API and a
  new decision entry.

## Unresolved and excluded behavior

No known material ambiguity in the current public surface is unresolved.
Format auto-negotiation, arbitrary charset guessing, XML canonicalization,
DTD/entity expansion, general YAML graph preservation, MessagePack extension
registration, validation of incoming CBOR deterministic form, BSON scalar
encoding, HTTP policy, WSDL/XSD, schema validation, signatures, compression,
and application mapping are outside the current claim. Adding one requires a
new decision before runtime implementation.
