// Package wire provides shared format and error primitives for structured
// wire-format packages.
//
// Format-specific behavior lives in the jsonwire, xmlwire, soap, yamlwire,
// tomlwire, msgpackwire, cborwire, and bsonwire packages. Detection is
// deliberately limited to opt-in JSON/XML inspection because callers at
// interoperability boundaries usually know which wire format a peer promises
// to send.
//
// Start with an explicit format package; DetectFormat is only a leading-byte hint.
// Functions own no transport, context, handles or workers. Keep borrowed inputs
// stable; codecs may mutate decode targets before returning an error. Finite input
// and output quotas are not universal heap, callback-work or deadline guarantees.
// Default Error text is categorical; causes and structured diagnostics are sensitive.
package wire
