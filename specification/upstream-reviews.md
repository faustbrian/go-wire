# Upstream authority review history

This append-only record preserves reviewed changes to the authorities monitored
by [`monitoring.json`](monitoring.json). A monitoring digest changes only after
the corresponding upstream delta has been classified against the applicable
specification decisions.

## 2026-09-09: RFC 9110 errata

- **Authority:** `rfc9110-errata`
- **URL:** https://errata.rfc-editor.org/search/?rfc_number=9110&presentation=records
- **Previous SHA-256:**
  `1f6790054c0cdb2f2a70a94fa2b9c73b09a4ee0578a32b4a3006ed0ecfaac86d`
- **Reviewed SHA-256:**
  `cec32fd170146656d933f627b512f2e027ae5c3592f5ec7760c3627493b30505`
- **Retrieved and reviewed:** 2026-09-09
- **Applicability:** `WIRE-DEC-016` directly; `WIRE-DEC-004` as source
  applicability only
- **Disposition:** Behavior-neutral because HTTP grammar and transport remain
  caller-owned policy.

Three consecutive retrievals produced the reviewed digest. The authority now
includes [Errata ID 9164](https://errata.rfc-editor.org/eid9164/), reported on
2026-09-07 as a Technical erratum against RFC 9110 Appendix A. It proposes
documenting that the collected ABNF normalizes equivalent rule spellings in
addition to the already documented list-rule expansion. The report expressly
states that the collected grammar and body define the same language, so it
does not identify an interoperability change.

`WIRE-DEC-016` leaves HTTP grammar and transport behavior to composed callers,
and `WIRE-DEC-004` neither parses HTTP ABNF nor negotiates media types. Errata
ID 9162 remains Reported, and the immutable RFC 9110 source and its selected
digest are unchanged. No selected wire behavior or source binding changes.
Reconsider this disposition if either erratum becomes Verified or if the
package later owns HTTP grammar or field combination.

## 2026-09-06: Go releases feed

- **Authority:** `go-releases`
- **URL:** https://go.dev/dl/?mode=json&include=all
- **Previous SHA-256:**
  `638127a053a86576fc235aa196b26145c6f2fce8ce839ded767212a18d1c9415`
- **Reviewed SHA-256:**
  `1ed915f72633d0a72eaa2f462740153db4fe347cb56f7d1e25ec44868568f13e`
- **Retrieved and reviewed:** 2026-09-06
- **Applicability:** `WIRE-DEC-001`, `WIRE-DEC-014`, and `WIRE-DEC-015`
- **Disposition:** Behavior-neutral for the selected Go 1.26.6 contracts.

Three consecutive retrievals produced the reviewed digest. The release
inventory still begins with Go 1.27.1 and Go 1.26.8, so no newer release was
added after the previous review. The prior response body was not retained, so
the byte-level representation change cannot be reconstructed. This feed is a
release-discovery authority: the immutable Go 1.26.6 JSON, XML, errors, and
language source URLs and their selected digests remain unchanged. No newer Go
behavior is adopted by this monitoring refresh.

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

## 2026-09-03: CTAP releases feed

- **Authority:** `ctap-releases`
- **URL:** https://fidoalliance.org/feed/
- **Previous SHA-256:**
  `cfcdcbebc4149e0a2a10a11e3c428c7e5819c3696a89905b932c8fd943f63adf`
- **Reviewed SHA-256:**
  `0a4f5ac58a20f05d92074b4222de33736d38e40e4746bc7f50c9bf2469836989`
- **Retrieved and reviewed:** 2026-09-03
- **Applicability:** `WIRE-DEC-011` directly; `WIRE-DEC-001` as the broader
  codec-delegation source inventory
- **Disposition:** Behavior-neutral for the selected CTAP 2.2 deterministic
  CBOR profile.

The feed published one item after the previous 2026-08-30 review:
[WebAuthn Level 3 became a W3C Recommendation](https://fidoalliance.org/webauthn-level-3-is-now-a-w3c-recommendation/)
on 2026-08-31. The announcement mentions CTAP 2.2 and CTAP 2.3 as paired
authenticator-protocol versions, but it does not revise CTAP 2.2 or its
deterministic-CBOR rules. This is a publication-date bounded semantic
comparison because the previous feed response body is not available for a
byte-for-byte comparison.

The immutable [CTAP 2.2 Proposed Standard dated
2025-07-14](https://fidoalliance.org/specs/fido-v2.2-ps-20250714/fido-client-to-authenticator-protocol-v2.2-ps-20250714.pdf)
remains the selected normative source with SHA-256
`2ef853c63fd0835e609e16fbe6a7fcd5965fd6a21d86752f46df54c5cfb81f73`.
The [specification index](https://fidoalliance.org/specifications/download/)
also lists CTAP 2.3 Proposed Standard dated 2026-02-26 and CTAP 2.3.1 Working
Draft dated 2026-05-29; both predate the previous review and neither changes
the selected source binding. A future maintenance change should replace the
general-news feed with a specification-specific release authority so unrelated
announcements do not create authority alerts.

## 2026-09-03: RFC 9110 errata

- **Authority:** `rfc9110-errata`
- **URL:** https://errata.rfc-editor.org/search/?rfc_number=9110&presentation=records
- **Previous SHA-256:**
  `38bd006c96f8963d58573f704c5313a5f81968b90738c03ade0b036ec7bbdf4b`
- **Reviewed SHA-256:**
  `1f6790054c0cdb2f2a70a94fa2b9c73b09a4ee0578a32b4a3006ed0ecfaac86d`
- **Retrieved and reviewed:** 2026-09-03
- **Applicability:** `WIRE-DEC-016` directly; `WIRE-DEC-004` as source
  applicability only
- **Disposition:** Behavior-neutral because HTTP field combination remains
  caller-owned transport policy.

[Errata ID 9162](https://errata.rfc-editor.org/eid9162/) was reported on
2026-09-01 as a Technical erratum against RFC 9110 Section 5.2. It proposes
changing the repeated-field combination wording from values separated by a
comma to values separated by comma plus space so the rule matches its example.
The erratum remains Reported, not Verified, and does not revise the immutable
RFC 9110 source.

`WIRE-DEC-016` expressly leaves HTTP headers and transport behavior to composed
callers. `WIRE-DEC-004` examines leading payload bytes and neither consumes
HTTP fields nor negotiates media types. No selected behavior or source binding
changes. Reconsider this disposition if Errata ID 9162 becomes Verified or if
the package later owns HTTP field combination.
