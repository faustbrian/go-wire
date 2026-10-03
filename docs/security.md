# Security

## Trust Boundary

Treat all JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON input as
untrusted. Choose explicit decode limits and reject unsupported data shapes
before application processing.

## Format Risks

BSON raw validation uses iterative traversal with a fixed 100-level nesting
limit below the root, covering documents, arrays, and CodeWithScope scopes.
Structural checks remain enabled with duplicate-key opt-in. Driver validation
is shallow, so the package validates every entered container before decoding.
Post-encoding validation does not bound driver buffering or custom codec work;
pre-encoding allocation protection remains separate work.

XML and SOAP require entity and nesting discipline. YAML aliases and recursive
structures require limits. Binary formats require byte, depth, collection, and
allocation bounds. Never assume equivalent semantics across formats.

## Application Responsibilities

MessagePack preflight retains recursive raw-key duplicate rejection. Its
additional destination-projection check covers built-in, unregistered scalar
and array keys in known maps, nested arrays/slices, and unambiguous direct or
`noinline` struct fields. It preserves numeric-width normalization and explicit
`AllowDuplicateKeys` behavior; it is not a general driver emulator.

Custom decoder/unmarshaler interfaces are not invoked by projection preflight.
Opaque `msgpack.Register` overrides cannot be detected through the driver's
public API. Applications own these trusted collaborators and their key
normalization policy: use a codec-specific duplicate check or explicit
last-key-wins policy, and review it whenever registration or codec behavior
changes. This limitation avoids executing callbacks twice; it does not prove
that arbitrary custom projections preserve nominal key equality.

Wire maintainers still own projection gaps for struct and non-empty interface
keys, alias, automatic
embedding, shadowed or interned field routing, and existing dynamic interface
destinations. Raw-key and numeric-fit checks remain active, but these surfaces
are not certified for destination-projected uniqueness. Review these gaps
before expanding the supported projection contract or the next major release.

MessagePack's per-decode admission owner counts every container, scalar, key
and value once before generic materialization. `MaxTotalValues` defaults to
256 Ki values. `MaxKeyComparisonWork` defaults to 8 Mi abstract examined-byte
or value units. Both bounds are inclusive; zero selects the finite default,
and negative options are rejected before reading.

For each map with `n` keys, a key weight is its exact encoded byte span plus
four times its subtree value count. Before generic decoding, the owner
reserves two passes of `n*(n-1)/2 + (n-1)*sum(key weights)`. Arithmetic is
checked against the remaining allowance before multiplication. Duplicate-key
opt-in skips this comparison reservation, but never structural admission.
This conservative reservation bounds admitted comparison inputs and weighted
pairwise work; it does not establish a wall-clock CPU or actual heap quota.

Known built-in array-key projections additionally reserve expanded target
preparation and comparison through the same owner, before reflected arrays
or projection tuples are allocated. For a scalar leaf, storage/preparation
weights are one and comparison weight is zero. An array of `n` elements has
storage `n*child storage`, preparation `storage+n+n*child preparation`, and
comparison weight `n+n*child comparison`. A map reserves preparation for
each key and comparison weight for each key's possible peers, including zero
padding. Byte arrays use this same conservative estimate even when no tuple
is needed. This estimate does not emulate opaque custom keys.
Array components must themselves be supported built-in scalars or arrays;
pointer, struct, interface and custom components are not covered by expanded
preparation accounting.
Projection preflight declines unsupported shapes before preparing reflected
values or tuples; their final driver allocations remain a separate boundary.

Wire maintainers still own final driver destination allocations, unsupported
or ambiguous projection shapes, and context-aware reader/decode APIs. Custom
objects and callback work remain trusted collaborator responsibilities.
Review these residual boundaries before claiming aggregate memory, CPU-time,
or cancellation protection beyond the admitted preflight operations.

The new option fields and finite aggregate admission are pending next-major
changes. Callers using unkeyed `DecodeOptions` literals must migrate to keyed
literals; named limit overrides may be needed for accepted larger workloads.

Ordinary `wire.Error` and `soap.FaultError` text exposes classifications only
in the pending major-release source. Structured fields and wrapped causes are
retained for trusted inspection, not automatic logging or disclosure. Explicit
field dumps, formatting underlying causes, and application-added error prefixes
require application redaction. This is not a guarantee about arbitrary caller
formatting or caller-owned error implementations.

Apply transport body limits, deadlines, authentication, authorization, and
rate limits before decoding. Do not expose raw parser errors when they may
contain sensitive input.

See [dependencies](dependencies.md), [formats](formats.md), and
[hardening](hardening.md) for the maintained boundary evidence. Report
vulnerabilities through [SECURITY.md](../SECURITY.md).
