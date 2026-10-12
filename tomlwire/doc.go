// Package tomlwire provides bounded, strict TOML document decoding and
// deterministic encoding.
//
// TOML duplicate keys and malformed trailing text are always rejected.
// Datetime values use time.Time and TOML local date/time types according to
// the underlying decoder. Numeric conversions reject overflow and precision
// loss. TOML is not included in wire.DetectFormat because arbitrary text
// cannot be identified reliably.
//
// Use Decode/Encode for the pinned TOML 1.1.0 codec contract; an empty document
// is valid. Reader calls buffer the document and never close the reader. Inputs,
// custom codecs and mutable destinations are caller-owned; a decode failure can
// leave assignments in place. No context API, transport, cgo or workers are owned.
package tomlwire
