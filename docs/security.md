# Security

## Trust Boundary

Treat all JSON, XML, SOAP, YAML, TOML, MessagePack, CBOR, and BSON input as
untrusted. Choose explicit decode limits and reject unsupported data shapes
before application processing.

## Format Risks

BSON raw validation uses iterative traversal with a fixed 100-level nesting
limit below the root, covering documents, arrays, and CodeWithScope scopes.
Structural checks remain enabled with duplicate-key opt-in. Driver validation
is shallow, so the package validates every entered container before decoding.
Post-encoding validation does not bound driver buffering or custom codec work;
pre-encoding allocation protection remains separate work.

XML and SOAP require entity and nesting discipline. YAML aliases and recursive
structures require limits. Binary formats require byte, depth, collection, and
allocation bounds. Never assume equivalent semantics across formats.

## Application Responsibilities

Ordinary `wire.Error` and `soap.FaultError` text exposes classifications only
in the pending major-release source. Structured fields and wrapped causes are
retained for trusted inspection, not automatic logging or disclosure. Explicit
field dumps, formatting underlying causes, and application-added error prefixes
require application redaction. This is not a guarantee about arbitrary caller
formatting or caller-owned error implementations.

Apply transport body limits, deadlines, authentication, authorization, and
rate limits before decoding. Do not expose raw parser errors when they may
contain sensitive input.

See [dependencies](dependencies.md), [formats](formats.md), and
[hardening](hardening.md) for the maintained boundary evidence. Report
vulnerabilities through [SECURITY.md](../SECURITY.md).
