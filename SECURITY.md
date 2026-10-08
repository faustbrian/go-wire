# Security Policy

## Supported Versions

The latest published stable major is
[`github.com/faustbrian/go-wire/v3` v3.0.1](https://github.com/faustbrian/go-wire/releases/tag/v3.0.1).
V2.0.0 remains available at its separate module path, without v3's narrowed
MessagePack admission and projection contract. The original v1 default-error
rendering is not the categorical error contract introduced in v2; do not
assume a v1 backport. Review the
[versioned threat model](docs/security-threat-model-v1.md) and
[migration guidance](docs/migration.md).

## Reporting A Vulnerability

Use GitHub private vulnerability reporting for this repository. Include a
minimal reproducer, expected and observed behavior, affected versions, impact,
and any suggested mitigation. Do not include secrets or production data.

## Response Process

Maintainers will acknowledge the report, reproduce and assess it privately,
coordinate a fix and advisory, and credit the reporter when requested. Public
disclosure should wait until a fix or agreed mitigation is available.

## Package Security Boundary

All supported text and binary formats are untrusted parser inputs. Byte, depth, collection, alias, extension, and output limits are part of the maintained security boundary.

## Application Responsibilities

Applications remain responsible for transport limits, authentication,
authorization, rate limiting, deadlines, secret handling, deployment policy,
and business-level validation. Package safeguards do not replace those
controls.

See [docs/security.md](docs/security.md) and the
[current model](docs/security-threat-model-v1.md) for adoption boundaries.
[docs/hardening.md](docs/hardening.md) is a historical pre-v1 audit, not
current-release qualification.
