# Security policy

## Supported versions

Before v1, only the latest commit on `main` receives security fixes. After v1,
the latest minor line will be supported; additional maintained lines will be
listed here if the policy changes.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting feature for this repository. Do
not open a public issue containing exploit details, sensitive payloads, or a
zero-day report.

Include:

- affected version or commit;
- format and API involved;
- a minimal redacted reproducer;
- security impact and preconditions;
- suggested mitigation, if known.

Maintainers should acknowledge a report within seven days, coordinate a fix and
advisory privately, and credit the reporter unless anonymity is requested.

## Security boundaries

`go-wire` parses attacker-controlled bytes but does not provide authentication,
authorization, transport security, schema validation, XML signatures, secret
redaction, or safe logging policy. Callers must set appropriate payload limits
and avoid logging raw payloads containing sensitive data.
