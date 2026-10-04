# Repository Standards

This repository follows the shared maintenance baseline used by the
`faustbrian/go-*` OSS packages.

## Mandatory Root Files

Every repository contains `.gitattributes`, `.gitignore`,
`.golangci.yml`, `AGENTS.md`, `CHANGELOG.md`, `CLAUDE.md`,
`CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `LICENSE`, `Makefile`, `NOTICE`,
`README.md`, `ROADMAP.md`, `SECURITY.md`, `THIRD_PARTY_NOTICES.md`,
`llms.txt`, and `llms-full.txt`.

AI planning and execution briefs live in `.ai/GOAL.md` and
`.ai/GOAL_HARDEN.md`, keeping internal agent material separate from the
package's public documentation surface.

`NOTICE` identifies the project and its ownership. `THIRD_PARTY_NOTICES.md`
separately records attribution and provenance for copied, forked, generated,
or vendored third-party source. Both remain present even when no third-party
source requires attribution. A package may retain a different approved OSS
license when provenance requires it.

## Mandatory Documentation

The shared taxonomy is lowercase kebab-case and includes a documentation
index, quickstart, adoption guide, API reference, architecture, examples,
cookbook, FAQ, troubleshooting, migration, compatibility, performance,
hardening, security, Go safety and concurrency, and releasing guide.
Package-specific documents extend this taxonomy without renaming shared
concepts.

## Mandatory Automation

Wire uses an immutable reusable CI workflow for selected-module quality,
benchmark, fuzz, and security gates. The supported minimum is Go 1.27.0,
as declared in `go.mod` and the module manifest. Dependency review runs on
pull requests; reachable dependency scanning uses `govulncheck`.

The current Make interface is `check`, `ci`, `cohesion`, `inventory`, and
`repository-check`. Formatting, race, coverage, analyzer, fuzz, benchmark,
and vulnerability checks are gates inside the library tool, not separate
Make targets. A release rehearsal uses the manual CI input
`release_dry_run`; tags do not automatically publish releases. Follow the
[release guide](releasing.md) for guarded tagging and manual publication.

The package family shares the `GO-SAFETY-1` baseline. It forbids `unsafe`,
cgo, and `go:linkname` in production code and standardizes ownership,
goroutine lifecycle, race, fuzz, resource-bound, leak, and benchmark evidence.

## Approved Package-Specific Differences

- `jsonapi` carries JSON:API feature, conformance, extension/profile,
  recommendation, and threat-model documentation.
- `jsonrpc` carries protocol conformance and middleware documentation.
- `queue` carries backend, delivery, lifecycle, failure, and integration
  documentation plus a live-backend integration workflow. Its fork provenance
  requires detailed third-party notices.
- `wire` carries format, dependency, and audit-evidence documentation.
- `tabular` carries format and ingest-limit documentation. It uses
  Apache-2.0 and retains XLS provenance notices.

Code, dependencies, fuzz targets, benchmark inputs, and domain-specific
security guidance are expected to differ. Shared policy wording and automation
structure must not drift without updating this contract across all affected
repositories.
