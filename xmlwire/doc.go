// Package xmlwire provides bounded XML encoding and decoding with explicit
// namespace, root-element, strictness, and character-set policy.
//
// Use Decode or Root for XML 1.0/namespace-aware token validation, then Encode
// for package-selected XML, not canonical XML or schema validation. Decode uses
// independent validation and target passes; charset callbacks may run twice.
// Directives are not DTD validation or external entity acquisition. Reader calls
// buffer input and never close it. Targets/custom callbacks are caller-owned;
// no context API, transport, cgo or background workers are introduced.
package xmlwire
