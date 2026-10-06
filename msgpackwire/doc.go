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
// not covered by that additional projection check. Custom codecs and global
// msgpack.Register overrides are trusted caller-owned behavior, not predicted
// or executed by preflight. See docs/security.md for remaining boundaries.
//
// MaxTotalValues admits aggregate structure before generic materialization.
// MaxKeyComparisonWork reserves conservative raw and projected comparison work
// and supported expanded array-key preparation before those operations. Zero
// options retain finite defaults. These quotas do not bound final driver
// allocations, custom codec objects, elapsed CPU time, or reader cancellation.
package msgpackwire
