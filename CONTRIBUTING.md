# Contributing

## Before Editing

1. Read the README, [`AGENTS.md`](AGENTS.md) and the affected API contract.
2. Start with `go test ./...` or a focused package/test selector.
3. Inspect direct consumers when changing a public or composition contract.
4. Preserve unrelated work and generated/corpus provenance.

An ordinary fix stays in this repository. It does not require sibling checkouts,
Tools edits, ecosystem catalog updates or agent-goal ledgers. Change manifests
only when their owned facts change, not merely because source changed.

## Changes

Keep commits focused and conventional. Update changelogs for material public
behavior, compatibility or adoption changes, not CI-only administration.
Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected entries in the
[specification decision register](docs/specification-decisions.md), and
complete the Specification Decisions section of the pull request template. An
unresolved interpretation or stale source pin is release-blocking; peer
behavior cannot silently select policy.

Mutation is optional developer-local or explicitly selected full assurance, not a PR
prerequisite. Investigate viable survivors and add focused regressions for real
defects; moving a campaign off PRs does not waive a confirmed correctness or
security defect. Equivalent/invalid mutants retain reviewed narrow dispositions.

Do not fork shared verification workflows, add permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

For ordinary source work, `go test ./...` runs deterministic tests, fuzz seeds
and saved corpus regressions. With the qualified verifier selected by repository
policy, use:

```bash
make inventory
make check
make ci
```

`check` and `ci` use bounded local checks, not full mutation, conformance or
benchmark campaigns. CI consumes the checksum-pinned public Tools binary;
it does not build the verifier from source for ordinary contributions.
Documentation changes need relevant structure, links and examples, not unrelated
runtime work. Public API, concurrency and external-boundary changes retain their
affected compatibility, race or integration checks.

For active exploration, select one fuzz target and a small budget; long campaigns
remain explicit full assurance. Optional local mutation selects one package with
the qualified verifier. These examples are not all-change requirements:

```bash
GOMAXPROCS=2 go test -run '^$' -fuzz '^FuzzDecode$' -fuzztime=2s -parallel=1 -timeout=2m ./jsonwire
golib mutation --module . --package ./jsonwire
```

`make check-full` retains explicit full assurance; relevant release qualification
uses `golib release dry-run --module .`. The local maintainer release script
also runs `check-full` before tagging; see [release guidance](docs/releasing.md).
Select full or release checks for their material risk or release boundary, not
every commit. Report unavailable/failed checks honestly; a fast local pass is not
public-release or hosted-CI qualification.

## Selected Monthly Assurance

The owner-level [selected-assurance caller](.github/workflows/assurance.yml) is
configured for main only, on day 22 of each month at 08:17 UTC. It selects one
of the 9 declared root-module package-target identities using the shared
UTC-week selection, with 30 seconds of active exploration. Monthly selection
samples one identity; it does not promise exhaustive rotation. The selected
risks are cross-format round trips and malformed decoder input across JSON,
XML, SOAP, YAML, TOML, MessagePack, CBOR and BSON. The seven package-owned `FuzzDecode`
identities remain distinct; this sample does not certify every format or replace
independent interoperability, conformance and release qualification.

Base-free ordinary checks execute deterministic seeds, not active exploration;
committed-range CI selects affected active smoke. Manual assurance requires a
declared selector from the root-module fuzz operation in `.golib.yaml`, for
example `./jsonwire::^FuzzDecode$`. A bare anchored target is
accepted only when unique across declared packages; ambiguous, empty, unknown
and malformed selections refuse. The shared job bounds setup and execution
to ten minutes, active workers and CPU to two, owned memory to 2 GiB, native
output to 1 MiB and private-cache lifetime. Failure, cancellation, exhausted
budgets and cleanup refusal remain attributable failures. This lane does not
run mutation or replace Required, original full fuzz/race budgets,
explicit `make check-full`, or relevant public-consumer/release qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.

## Source Documentation

Install the checksum-pinned public verifier declared in `.golib.yaml`.
`make source-docs-check` validates objective Go comment, marker, link and
generated-producer structure; `make check`, `make check-full` and `make ci`
require it before their existing checks. Shared CI runs the same strict
contract and includes its result in Required. This is structural enforcement,
not a substitute for source-truth review or behavioral verification.
