# Migration notes

## From v2 to v3

Published version 3 uses `github.com/faustbrian/go-wire/v3`. Update
the root and all codec imports together. Error types and sentinels from different
major modules have distinct Go identities. Historical v1/v2 API inventories
remain retained; no consumer is automatically migrated.

Use keyed `msgpackwire.DecodeOptions` literals for the new `MaxTotalValues`
and `MaxKeyComparisonWork` fields. Zero selects finite inclusive defaults:
256 Ki aggregate values and 8 Mi conservative key-work units. Refusal occurs
before destination mutation or destination codec callbacks. Supported built-in
key projections also reject destination collisions; explicit duplicate-key
opt-in retains its documented behavior and aggregate node limits. Custom
codecs, ambiguous routing, final driver allocation and cancellation are
separate boundaries. Set deliberately reviewed finite limits for larger valid
inputs rather than assuming the v2 accepted-input set is unchanged.

Public APIs exposing Wire types require an explicit compatibility decision
before adopting a different major. Private integrations and test harnesses do
not require unrelated SDK majors solely for their internal import update.

## From v1 to v2

Use `go get github.com/faustbrian/go-wire/v2@v2.0.0` and update the root and
all codec imports to `github.com/faustbrian/go-wire/v2`. Update related imports
in one change: v1 and v2 errors and sentinels have distinct Go identities.
Existing v1 consumers remain on their selected version until deliberately
migrated; this release does not migrate other repositories.

Ordinary wire and SOAP fault text is categorical. Inspect `*wire.Error`, its
wrapped cause, and `*soap.FaultError` fields for trusted diagnostics instead of
matching old diagnostic strings. Existing classification and cause inspection
remain available within the selected major module.

Built-in XML and SOAP charset conversion enforces raw-input byte quotas and
rejects unsupported or overlong labels before reading. Use
`xmlwire.CharsetReaderWithLimit` for an explicit quota; custom callbacks and
reader responsiveness remain caller-owned. Conversion can expand output.

BSON validates nested documents, arrays and CodeWithScope scopes before decoder
callbacks, including with duplicate-key opt-in. Supply consecutive array indices
and at most 100 nested containers below the root. Encoding validates structure
after codec work; this is not pre-encoding allocation protection.

YAML scalar values, explicit mapping keys and quoted continuations retain their
authored content. Emitted block indicators can change to preserve those values;
encoded bytes are not promised to be identical across released versions.

## Version 2 compatibility

The module follows stable v2 compatibility. Pin a released version, keep the
integration behind boundary adapters, and review `CHANGELOG.md` whenever updating.

## Published v2 charset boundary

Version 2 charset hardening changes admission and default quotas, not
a partial v1 publication. Direct `xmlwire.CharsetReader` calls now accept at
most 1 MiB of raw input and labels at most 64 bytes, and return categorical
diagnostics. Use `CharsetReaderWithLimit` for a deliberate larger finite raw
input quota; zero selects the default and negative limits are invalid.
Use `errors.Is` to inspect retained reader causes rather than parsing text.
XML and SOAP built-in callbacks follow their configured raw-input quotas;
UTF-8 conversion may expand that input up to threefold. Custom callbacks and
reader cancellation remain application responsibilities.

## Published v2 BSON structure boundary

BSON validation now includes CodeWithScope scope documents and all nested
document/array boundaries, including when duplicate keys are explicitly allowed.
Arrays require consecutive zero-based indices. Raw nesting is limited to 100
containers below the root; documents, arrays, and scopes each add one level.
Malformed structure fails before decoder callbacks. This acceptance change is
part of version 2. Encoding validates serialized output afterward;
driver buffering, pre-encoding resource limits, and custom-codec work are separate
boundaries, not protected by this raw traversal limit.

## From direct `encoding/json`

Before:

```go
err := json.NewDecoder(body).Decode(&response)
```

After:

```go
err := jsonwire.DecodeReader(body, &response, jsonwire.DecodeOptions{
	MaxBytes: 1 << 20,
})
```

Behavior differences:

- the stream is bounded;
- exactly one JSON value is required;
- invalid UTF-8 is rejected instead of being replaced with U+FFFD;
- invalid targets use `wire.ErrTarget`, byte limits use `wire.ErrSizeLimit`,
  and type mismatches retain structured validation errors;
- unknown fields remain accepted unless explicitly rejected;
- the default decoder still follows `encoding/json` duplicate-key behavior.

If existing code consumes a sequence of JSON values, do not migrate it to
`DecodeReader`; the helper intentionally rejects that protocol.

## From direct `encoding/xml`

Before:

```go
err := xml.NewDecoder(body).Decode(&shipment)
```

After:

```go
err := xmlwire.DecodeReader(body, &shipment, xmlwire.DecodeOptions{
	ExpectedRoot: xml.Name{Space: "urn:vendor", Local: "Shipment"},
})
```

Behavior differences:

- the stream is bounded;
- strict parsing is explicit and the default;
- the complete document must have one root;
- common legacy charsets are converted through an auditable allowlist;
- root validation uses resolved namespaces.

If an old decoder set `Strict = false`, capture malformed fixtures and enable
`AllowNonStrict` only after verifying the recovered values match current
production behavior.

## From hand-written SOAP structs

Common legacy code decodes an Envelope struct and checks a body field manually.
Migrate in two stages:

1. Parse with `soap.Parse` or `soap.ParseReader` and preserve current body DTOs.
2. Replace manual fault checks with `wire.ErrSOAPFault` and `*soap.FaultError`.

```go
envelope, err := soap.ParseReader(body, soap.ParseOptions{})
if errors.Is(err, wire.ErrSOAPFault) {
	return mapFault(envelope.Fault)
}
if err != nil {
	return err
}
return envelope.DecodeBody(&response)
```

Audit these differences:

- SOAP version comes from the Envelope namespace;
- Header must occur at most once before Body;
- Body is required exactly once;
- a Fault must be the only Body child and contain code and reason;
- `DecodeBody` expects one direct body child;
- SOAP faults are returned as errors together with their envelope.

## From direct YAML libraries

Replace unbounded decoding with `yamlwire.DecodeReader` and select compatibility
options deliberately. Duplicate keys and multiple documents are rejected by
default. Aliases and merge keys are accepted under bounded library protections
and can be rejected explicitly. Encoding is deterministic and supports an
explicit indentation width.

## From direct TOML libraries

Use `tomlwire.DecodeReader` for bounded, single-document decoding and
`tomlwire.EncodeWriter` for output. Unknown fields are optional strictness;
duplicate keys, trailing documents, numeric overflow, and incompatible datetime
targets are rejected. Verify emitted table layout if existing golden files
depend on a library-specific presentation.

## From direct MessagePack libraries

Use `msgpackwire.DecodeReader` and `msgpackwire.EncodeWriter`. The wrapper
requires exactly one object, rejects unknown extensions and lossy numeric
conversion, rejects duplicate keys, and defaults untyped maps to string keys.
Use `AllowDuplicateKeys` only for a measured legacy peer, where decoding uses
last-key-wins behavior. Preserve numeric widths when the protocol requires
them, or opt into normalization explicitly. Compare golden bytes because
compact-number and map-order settings affect output.
The safe defaults reject nesting beyond 32 levels, arrays beyond 131,072
elements, and maps beyond 65,536 pairs; set the corresponding `DecodeOptions`
fields only from measured peer requirements.

## From direct CBOR libraries

Choose `cborwire.Canonical`, `cborwire.CoreDeterministic`, or
`cborwire.CTAP2Deterministic` explicitly. Tags and indefinite-length items are
rejected by default, resource limits are bounded, and exactly one item is
required. Record the selected profile and time-tag behavior as part of the peer
contract before replacing existing emitted bytes.

## From direct BSON libraries

`bsonwire.Decode` accepts exactly one document and rejects duplicate keys
recursively. Object IDs, dates, ordered documents, raw documents, and arrays
remain available through the package aliases. Use `bsonwire.D`, a struct, or
raw bytes for stable ordering; `bsonwire.M` intentionally makes no order
guarantee. Integer-width minimization and lossy double-to-integer truncation
are opt-in compatibility choices.

## Error mapping

Published v2/v3 behavior: `wire.Error.Error()` renders only classification,
and `soap.FaultError.Error()` renders `soap fault`. Format, operation, original
causes, and SOAP fault fields remain inspectable; `errors.Is` and `errors.As`
contracts are unchanged. Replace assumptions about contextual error text before
adopting those major versions. Released v1 retains its existing text behavior.

Do not match message strings. Replace legacy string checks with `errors.Is` and
`errors.As`. Preserve the underlying cause only for redacted diagnostic logging and do
not return it to untrusted clients without review. Discard a decode target
after any error because reflection codecs may have assigned fields before a
later failure.

## Rollback

Keep the old boundary adapter available during the comparison window. A safe
rollback switches the adapter, not domain code. Record representative inputs
and output DTO comparisons before deleting the old implementation.

## Breaking releases after v1

Major upgrades will document:

- removed or renamed symbols;
- changed defaults or limits;
- error-classification changes;
- emitted-wire changes;
- normalization changes;
- a before/after migration example.

Never infer migration requirements only from compiler errors; wire behavior can
change without a signature change.
