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

All readers are byte-bounded. Safe defaults require exactly one top-level value
or document and reject ambiguous or resource-amplifying features:

- JSON invalid UTF-8, malformed syntax, trailing values, and optionally unknown
  fields; duplicate names retain documented `encoding/json` behavior;
- XML and SOAP nesting beyond the configurable 1,000-element default;
- YAML duplicate keys, excessive alias expansion, deep nesting, and multiple
  documents; aliases and merge keys remain bounded and can be rejected
  explicitly;
- TOML duplicate keys, trailing documents, overflow, and invalid datetime
  conversion;
- MessagePack impossible or excessive collection lengths, nesting beyond 32
  levels, duplicate keys, unknown extensions, non-string untyped map keys,
  lossy numeric conversion, and trailing objects;
- CBOR duplicate map keys, tags, indefinite-length values, excessive nesting or
  collections, and trailing items;
- BSON invalid length prefixes, trailing bytes, non-document top levels,
  recursive duplicate keys, and lossy double-to-integer conversion.

Options that relax a safe default or admit larger resource use must be enabled
only from protocol evidence. They do not remove the overall byte limit.
Dependency selection, update policy, and residual parser risks are documented
in [`docs/DEPENDENCIES.md`](docs/DEPENDENCIES.md).

Decode targets are not transactional: discard the target after any error.
Encode values are caller-owned and the current APIs do not impose an output
byte quota; bound source collections when output size can be influenced by an
untrusted party. See [`docs/HARDENING.md`](docs/HARDENING.md) for the threat
model, findings, and per-format evidence matrix.

All typed encode paths preflight application values and reject cycles or more
than 1,000 traversed levels before invoking a recursive codec. Custom marshalers
remain responsible for terminating their own method bodies.
