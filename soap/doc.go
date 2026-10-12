// Package soap provides transport-neutral SOAP 1.1 and SOAP 1.2 envelope,
// body, and fault primitives while preserving access to the original XML.
//
// Use Parse/ParseReader for the SOAP 1.1 Note (2000-05-08) or SOAP 1.2 Part 1
// Second Edition (2007-04-27), then DecodeBody or inspect a FaultError. Envelope
// raw accessors copy bytes; exported fault fields remain mutable. Raw fragment
// well-formedness is not schema/header-role processing, WS-Security or signing.
// Callbacks/readers are borrowed and cannot be cancelled through this API; no
// HTTP client, handle closing, cgo or background workers are owned here.
package soap
