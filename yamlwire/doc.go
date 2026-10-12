// Package yamlwire provides bounded, explicit YAML 1.2 decoding and
// deterministic encoding.
//
// Decoding rejects duplicate keys and multiple documents by default. Anchors,
// aliases, and merge keys are supported with resource limits; callers can
// reject aliases or merge keys for protocols that do not permit them. YAML is
// not included in wire.DetectFormat because reliable detection from arbitrary
// text is not possible.
//
// Use Decode/Encode for the pinned YAML v4 mapping of YAML 1.2.2 with retained
// dependency compatibility rules, not cryptographic canonicalization. Reader
// calls buffer the stream and never close it. Inputs and custom codecs are
// borrowed, and target decoding can mutate before failure; serialize shared
// mutation. No context API, transport, schema, cgo or workers are introduced.
package yamlwire
