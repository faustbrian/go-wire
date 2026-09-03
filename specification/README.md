# Specification conformance matrix

`manifest.tsv` pins the normative sources used by the wire decision register.
RFC sources use immutable RFC Editor text; W3C publications use dated URLs;
repository-hosted specifications use immutable source commits; CTAP uses the
dated FIDO Alliance publication PDF. The published BSON 1.1 page is versioned
by its content digest. SHA-256 digests make source drift explicit.

The module claims only the documented package profile for each format. A
format specification describes its wire model; Go target conversion,
dependency behavior, finite service limits, and package error classification
remain separate decisions. Passing a dependency's tests is not evidence that
the wrapper implements every optional feature of a format.

The canonical
[`docs/specification-decisions.md`](../docs/specification-decisions.md)
records every material interpretation, consequence, and condition for
reconsideration behind this conformance matrix.

## Upstream review history

The append-only [upstream authority review history](upstream-reviews.md)
records each monitored authority delta, its exact decision applicability, and
the disposition required before its reviewed digest changes.

| Decision | Conformance scope |
| --- | --- |
| WIRE-DEC-001 | Specification editions and codec delegation |
| WIRE-DEC-002 | Complete inputs and destination mutation |
| WIRE-DEC-003 | Input, structural, graph, and output limits |
| WIRE-DEC-004 | Conservative format detection |
| WIRE-DEC-005 | JSON interoperability policy |
| WIRE-DEC-006 | XML parsing and namespace policy |
| WIRE-DEC-007 | SOAP envelopes and faults |
| WIRE-DEC-008 | YAML schemas, graphs, and streams |
| WIRE-DEC-009 | TOML documents and conversion |
| WIRE-DEC-010 | MessagePack maps, extensions, and limits |
| WIRE-DEC-011 | CBOR deterministic profiles and decoding |
| WIRE-DEC-012 | BSON documents, order, and conversion |
| WIRE-DEC-013 | Per-format determinism claims |
| WIRE-DEC-014 | Error classification and disclosure |
| WIRE-DEC-015 | Go graph validation and codec differentials |
| WIRE-DEC-016 | Transport, schema, and application ownership |

Run the focused map and evidence check with `make conformance`. For an update,
verify provenance and digest, review errata and codec changes, update decisions
and behavioral tests, then change the manifest. A digest change alone MUST NOT
silently change behavior.
