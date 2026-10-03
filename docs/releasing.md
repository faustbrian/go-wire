# Releasing

## Preconditions

A release must originate from a clean, synchronized `main` branch. Update
`CHANGELOG.md` with one dated version section, confirm compatibility and
migration notes, and verify dependency/provenance changes.

## Verification

```sh
make check
```

Format-specific fuzz targets, 100% meaningful production coverage, API
inventory checks, documentation links, and vulnerability scanning are release
gates.

## Tagging and publication

The repository has no automatic tag-triggered publisher. Dispatch the owned
`ci.yml` workflow with `release_dry_run=true` and inspect every required result
for the actual candidate. Hosted gates do not publish a release.

For the local maintainer route, `scripts/release.sh major` checks a clean,
synchronized main and the dated changelog, runs `make check`, and creates an
annotated tag. Use the trusted maintainer signing key for publication,
verify the new remote tag target, and never move an existing release tag.

Publish changelog-derived notes with a deterministic source archive and checksum
bound to the verified main commit. Verify public asset bytes, the public Go
module and its checksum record, and a clean consumer after publication. A tag
push alone does not establish a successful release or consumer adoption.
