// Package cborwire provides bounded CBOR decoding and deterministic encoding.
//
// Duplicate keys, indefinite-length items, and tags are rejected by default.
// Tags and indefinite lengths require explicit interoperability options.
// Encoding defaults to RFC 7049 canonical form, with explicit Core
// Deterministic and CTAP2 profiles available. Decoder nesting and collection
// limits remain bounded even when the byte limit is raised. CBOR is excluded
// from wire.DetectFormat because arbitrary binary data cannot be identified
// reliably.
//
// Use Decode/Encode with an explicit profile where peer byte identity matters.
// The pinned authorities are RFC 7049 section 3.9, RFC 8949 core deterministic
// encoding and CTAP 2.2 section 8; these are different profiles, not generic signing
// policy. Reader calls buffer one bounded item and never close the reader. Borrowed
// values/custom codecs require caller synchronization; failed targets are
// indeterminate. No context API, transport, cgo or workers are introduced.
package cborwire
