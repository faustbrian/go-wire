# Wire threat model

- Model: `WIRE-THREAT-MODEL-1.0.0`
- Applies to: `github.com/faustbrian/go-wire/v3` v3.0.0 and v3.0.1
- Reviewed: 2026-10-09
- Owner: Wire maintainers

## Assets and trust boundaries

Wire protects format interpretation, application destination state, process
availability, and payload confidentiality at explicit codec boundaries.
Payload bytes, declared lengths, nested collections, keys, XML namespaces,
SOAP faults, YAML aliases, and codec options may originate outside the
application trust boundary. Destination types, readers, writers, registered
codecs, custom encoders, and application callbacks are caller-selected
collaborators, not passive untrusted bytes.

The module supplies no transport, network resolver, authentication,
authorization, credential store, replay policy, or cryptographic protocol.
It does not fetch XML entities or schema references. Applications own those
boundaries before codec invocation. No package-owned background worker or
global application cache is introduced. Third-party codec registration is
outside the package's ownership and must remain application-controlled.

## Current controls and their limits

All eight format families have explicit byte admission and output quotas.
Format-specific structural controls cover nesting, collections, aliases,
extensions, and unsupported shapes as described in [formats](formats.md).
Typed encoding checks cycles and depth before codec invocation. These
controls are not a universal wall-clock or actual heap-allocation quota.
Encoded values are caller-owned objects; custom codec work and dependency
buffering are not preempted by the final output-size check.

MessagePack admits aggregate containers, scalars, keys, and values before
generic materialization. `MaxTotalValues` defaults to 256 Ki values and
`MaxKeyComparisonWork` to 8 Mi conservative examined-byte/value units.
The bounds are inclusive; zero selects finite defaults and negative options
are invalid. Supported built-in array projections reserve their expanded
preparation and comparison work through the same owner. Admission refusal
does not invoke destination codecs or mutate the destination.

Raw duplicate-key checks and supported built-in destination projections
remain distinct. Custom codecs, opaque registration, unsupported key types,
and ambiguous field routing are not certified for projected uniqueness.
See [security guidance](security.md) for the exact supported shapes and
comparison reservation. No arbitrary custom-driver equivalence is claimed.

BSON raw validation iteratively checks entered containers with a fixed
100-level nesting bound below the root. JSON retains standard-library
duplicate-member behavior; applications requiring unique names must enforce
that policy before decoding. No cross-format canonical signing protocol is
provided. Decode failures may leave a target partially assigned after
admission; discard the target after any error unless a specific API promises
otherwise.

Default `wire.Error` and `soap.FaultError` strings are categorical in v2 and
v3. Structured fields and wrapped causes remain diagnostic-only and may
contain payload or application details. Explicit dumps and underlying-cause
formatting require caller redaction. Original v1 error rendering is a
different contract and must not be treated as a v1 backport of this policy.

Reader, writer, and codec APIs do not universally accept contexts. Byte and
structural limits do not interrupt a blocked collaborator or bound time spent
inside a dependency. Callers must supply deadline-aware I/O and bounded
custom codecs; do not use a detached timeout goroutine as a reclamation
guarantee. Wire does not claim asynchronous cancellation of decoding.

## Retained risks and owners

These scoped acceptance decisions retain the documented caller contract;
they do not certify an unsupported shape or close a new reproducible defect.

| ID | Severity | Owner | Rationale | Mitigation | Review condition |
| --- | --- | --- | --- | --- | --- |
| WIRE-RISK-001 | Medium | Wire maintainers and application integrator | Final driver allocation and encoding buffers are not an actual heap quota. | Keep finite byte, depth, collection and aggregate allowances; admit destination/value shapes and isolate process capacity for hostile workloads. | Revisit when changing a codec, allowance, destination contract, or after an observed allocation-bound violation. |
| WIRE-RISK-002 | Medium | Wire maintainers | Unsupported or ambiguous MessagePack projections cannot guarantee normalized key uniqueness. | Use supported built-in key/field shapes or an application-owned normalized duplicate check; raw structural checks remain active. | Revisit before expanding supported shapes, changing driver field routing, or claiming general projected uniqueness. |
| WIRE-RISK-003 | Medium | Application integrator | Custom codecs and global driver registration are trusted application code. | Control registration, require bounded collaborators, and review normalization and ownership without invoking callbacks twice. | Revisit whenever registration or custom codec behavior changes. |
| WIRE-RISK-004 | Medium | Wire maintainers and application integrator | Contextless I/O and dependency calls cannot be preemptively cancelled. | Supply deadline-aware readers/writers, finite input policies and bounded custom work; use process isolation where reclamation is required. | Revisit for a context-aware API, blocking collaborator incident, or stronger cancellation promise. |
| WIRE-RISK-005 | Medium | Application integrator | JSON duplicate-member compatibility can differ from a protocol's uniqueness policy. | Validate unique names before security-sensitive decoding; do not use ordinary normalization as canonical signing. | Revisit when selecting a signing or authorization protocol or changing duplicate policy. |
| WIRE-RISK-006 | Low | Application integrator | Diagnostic fields, wrapped causes and partially assigned targets retain data beyond default error strings. | Log classifications, redact explicit diagnostics and discard targets after failure. | Revisit after error-contract, target-ownership or logging changes. |
| WIRE-RISK-007 | Medium | Wire maintainers | Maintained third-party parsers remain supply-chain and semantic boundaries. | Pin reviewed dependencies and actions; run vulnerability, license, secret and workflow checks plus affected conformance and interoperability tests. | Revisit when dependencies, normative inputs, tool sources or scanner findings change. |

## Release and evidence boundary

The [v3.0.0 release](https://github.com/faustbrian/go-wire/releases/tag/v3.0.0)
binds source `74657780` and the narrowed MessagePack contract. Its
[exact-main CI](https://github.com/faustbrian/go-wire/actions/runs/37457870112)
and [actual release rehearsal](https://github.com/faustbrian/go-wire/actions/runs/37460280849)
passed. The retained clean public proxy/SumDB consumer verified its module
identity, eight runtime family imports, ordinary round trips, finite aggregate
refusal and preserved destination state. Signed public assets matched the
reviewed source archive. These results are not proof of the residual heap or
cancellation guarantees explicitly excluded above.

[V3.0.1](https://github.com/faustbrian/go-wire/releases/tag/v3.0.1) binds
`060791370b0035fcdfd61b3883e13b7d1b2bc1e6`; its changes from v3.0.0 are
workflow and release metadata, not codecs or runtime dependencies.
[Exact-main CI](https://github.com/faustbrian/go-wire/actions/runs/37764414872)
passed. A separate clean public proxy/SumDB consumer verified v3.0.1 at
that exact source, imported all eight runtime format families, passed ordinary
composition and finite aggregate-refusal checks with preserved destination
state, and passed vet. It also confirmed the legacy v1 SOAP default diagnostic
rendering and categorical v2/v3 rendering using an ordinary fault response.
The immutable v3.0.0 behavioral evidence remains applicable to unchanged
runtime inputs; the v3.0.1 consumer does not rerun the complete release gates
or certify excluded heap and cancellation guarantees.

Owned International, Localized, Measurement, Opening Hours, Temporal test
composition, WSDL v2 and Tools consumer adoption use the published v3 public
contract. WSDL's separate XSD v2 migration is not a Wire release prerequisite.
The [pre-v1 audit](hardening.md) remains historical: its fixed findings do not
establish affected released versions. Published-version disclosure requires
separate immutable-source applicability assessment, not inference from a
historical finding label.

Review this model after changes to format admission, projection, dependencies,
errors, I/O ownership, cancellation, release tooling, or public guarantees.
