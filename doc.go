// Package wire provides shared format and error primitives for structured
// wire-format packages.
//
// Format-specific behavior lives in the jsonwire, xmlwire, and soap packages.
// Detection is deliberately opt-in because callers at interoperability
// boundaries usually know which wire format a peer promises to send.
package wire
