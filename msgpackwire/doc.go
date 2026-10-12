// Package msgpackwire provides bounded MessagePack decoding and deterministic
// encoding.
//
// Decoding accepts exactly one object. Untyped maps require string keys while
// typed map targets may declare other comparable key types. Integer widths are
// preserved unless NormalizeNumericWidths is selected. The standard timestamp
// extension is supported; unknown extension IDs are rejected. A structural
// preflight rejects truncated objects and impossible collection lengths before
// target allocation. MessagePack is deliberately excluded from
// wire.DetectFormat because arbitrary binary bytes cannot be identified
// reliably.
//
// Raw duplicate keys are rejected recursively by default. Destination-key
// collisions are also checked for built-in, unregistered scalar and array key
// projections in maps and unambiguous direct or noinline struct fields.
// Struct keys and alias, auto-inline, shadowed or interned field routing are
// not covered by that additional projection check. Destination projection does
// not execute custom key codecs to predict equality; arbitrary custom codecs and
// global msgpack.Register overrides remain trusted caller-owned behavior.
// See docs/security.md for the generic decoding and projection boundaries.
//
// MaxTotalValues admits aggregate structure before generic materialization.
// MaxKeyComparisonWork reserves conservative raw and projected comparison work
// and supported expanded array-key preparation before those operations. Zero
// options retain finite defaults. These quotas do not bound final driver
// allocations, custom codec objects, elapsed CPU time, or reader cancellation.
//
// Use Decode/Encode for the MessagePack format pinned in specification/manifest.tsv.
// Reader calls buffer input and never close the reader. Keep input graphs and
// borrowed callbacks stable; failed destination decoding is not transactional.
// No transport, context API, cgo or background workers are owned here.
package msgpackwire
