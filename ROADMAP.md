# Roadmap

The roadmap protects scope as much as it describes future work. Items are not
promises or compatibility commitments until released.

## v1.0: JSON, XML, and SOAP foundation

- stabilize the current explicit package layout;
- validate APIs against multiple real, redacted vendor fixture families;
- maintain meaningful 100% production-code coverage;
- exercise fuzz targets over longer scheduled runs;
- publish benchmark baselines without promising fixed performance numbers;
- complete external review of error and normalization contracts;
- release only after the documented support matrices match verified behavior.

## After core stabilization

Potential work, ordered by demonstrated user need:

- additional XML charsets through optional adapters;
- WSDL integration guidance or a separate adapter package;
- helpers for common transport patterns that do not turn the core into an HTTP
  client framework;
- schema-generation guidance where output can remain explicit and auditable.

## Deferred formats

YAML, TOML, and MessagePack are intentionally deferred. A proposal must show:

- repeated production demand not solved well by existing focused packages;
- format-specific semantics that belong in this project;
- a dedicated API that does not weaken JSON/XML/SOAP boundaries;
- fixtures, fuzzing, benchmarks, documentation, and compatibility policy equal
  to the existing formats.

## Explicit non-goals

- generic all-codec interfaces;
- HTTP clients, retries, authentication, or service discovery;
- generic RPC frameworks;
- queues, persistence, or business mapping;
- hidden schema inference or normalization;
- WSDL client generation inside the core runtime package.
