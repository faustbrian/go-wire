# Upstream authority review history

This append-only record preserves reviewed changes to the authorities monitored
by [`monitoring.json`](monitoring.json). A monitoring digest changes only after
the corresponding upstream delta has been classified against the applicable
specification decisions.

## 2026-09-03: Go releases feed

- **Authority:** `go-releases`
- **URL:** https://go.dev/dl/?mode=json&include=all
- **Previous SHA-256:**
  `c0696d6e5708cce9644204513ee4ecb994f3fd748065c54b0890158c34631207`
- **Reviewed SHA-256:**
  `638127a053a86576fc235aa196b26145c6f2fce8ce839ded767212a18d1c9415`
- **Retrieved and reviewed:** 2026-09-03
- **Applicability:** `WIRE-DEC-001`, `WIRE-DEC-014`, and `WIRE-DEC-015`
- **Disposition:** Behavior-neutral for the selected Go 1.26.6 contracts.

The feed added Go 1.26.8 at commit
`c293dd49cbe25e1fe8d97d94a5cb618e7b6d831e` and Go 1.27.1 at commit
`862c888e612ac346c7c4d99c9392bdfd265f33b0`, both published on 2026-09-01.
The exact Go 1.26.6-to-1.26.8 comparison contains no change under
`encoding/json`, `encoding/xml`, `errors`, or `doc/go_spec.html`, so the
repository's four immutable Go 1.26.6 source URLs and hashes remain selected.

Go 1.27.1 includes `encoding/json` fixes for quoted `null` handling on
stringified fields and propagation of underlying `io.ErrUnexpectedEOF` from
streams. Those changes belong to the Go 1.27 line and do not alter the current
Go 1.26.6 contract. Adopting Go 1.27 would separately reopen `WIRE-DEC-001` and
`WIRE-DEC-015`; no changed `errors` source affects `WIRE-DEC-014` in the
reviewed Go 1.26 patch line.
