# Versioning and release guide

## Compatibility contract

The project uses Semantic Versioning. Beginning with v1.0.0, compatibility
includes more than Go signatures:

- exported packages, types, constants, variables, functions, methods, and
  struct fields;
- default payload limits and strictness;
- `errors.Is` and `errors.As` classifications;
- deterministic emitted bytes where documented;
- JSON normalization behavior;
- SOAP envelope and fault interpretation;
- YAML alias, merge, duplicate-key, and document-stream handling;
- TOML metadata, datetime, and numeric conversion behavior;
- MessagePack map-key, duplicate-key, extension, width, compact-encoding,
  nesting, and collection-limit behavior;
- CBOR deterministic profiles, tag policy, and resource limits;
- BSON document validation, duplicate-key handling, and numeric conversion;
- supported charset labels;
- documented accepted and rejected shapes.

Fixing behavior that contradicts a specification can still be breaking for
users. Release notes must call out the impact and migration path.

The [specification decision register](specification-decisions.md) records the
exact editions, ambiguities, dependency seams, selected behavior, consequences,
and executable evidence behind this compatibility surface. A changed parsing,
encoding, normalization, or error decision requires compatibility review even
when the previous behavior was undocumented. Normative source provenance is
pinned in [`../specification/manifest.tsv`](../specification/manifest.tsv).

## Version policy

- Patch releases contain compatible bug fixes, security fixes, documentation,
  and new regression fixtures.
- Minor releases add backward-compatible APIs or capabilities.
- Major releases can change or remove compatibility-governed behavior and must
  include migration notes.
- Breaking API changes require a new major release and remain documented.

## Deprecation policy

Prefer a documented deprecation in at least one minor release before removing
an exported API. A deprecation comment must name the replacement and migration
constraint. Security issues can require faster removal and must explain the
exception in release notes.

## Release prerequisites

The release commit must have:

- a clean worktree;
- the intended Go support matrix passing;
- 100% production statement coverage with behavior-focused tests;
- formatting, `go vet`, and lint gates passing;
- fuzz smoke targets passing;
- benchmark compilation and smoke execution passing;
- documentation links and examples validated;
- specification decisions, source pins, and conformance evidence validated;
- dependency and vulnerability scans passing;
- a dated `CHANGELOG.md` entry covering the release delta;
- migration notes for every breaking change.

## Release procedure

1. Choose a SemVer version from the compatibility impact and add a dated
   changelog entry covering the release delta. Keep an Unreleased section for
   subsequent work and document every breaking migration.
2. Merge the reviewed release preparation through the normal required gates.
3. Follow the [release guide](releasing.md) for the exact candidate rehearsal,
   guarded local tag creation, and manual publication procedure. The repository
   has no `make release-*` targets or automatic tag-triggered publisher.
4. Verify the published tag, release assets, public module checksum record,
   and clean-consumer resolution before reporting publication complete.

Never move or recreate a published tag. Publish a new patch version if a
release artifact or note needs correction.

## Release reproducibility

The release contains source only; consumers compile it with Go modules.
Publication binds the tag and commit to a deterministic source archive and
SHA-256 checksum. Runtime dependencies are pinned through `go.mod` and recorded
with their selection rationale and residual risks in
[`dependencies.md`](dependencies.md). Dependency upgrades require full wire
compatibility, fuzz, benchmark, vulnerability, and license review.

## Emergency security releases

Use GitHub's private vulnerability reporting flow. Prepare the fix and advisory
privately, run the complete gate, then publish the advisory, patch release, and
changelog entry together. Do not disclose exploit details in a public issue
before users can upgrade.
