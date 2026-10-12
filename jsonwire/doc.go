// Package jsonwire provides bounded, strict JSON helpers for interoperability
// boundaries where encoding/json's low-level defaults need explicit policy.
//
// Use Decode/Encode for RFC 8259 text; Normalize is presentation normalization,
// not a signing canonicalization profile. Reader calls buffer a complete bounded
// unit; they do not stream values or close the reader. Inputs/callbacks remain
// caller-owned and targets may be partially mutated on error. No context API,
// transport, schema, cgo or background workers are owned here.
package jsonwire
