# Contributing

Thank you for helping improve `go-wire`. Interoperability changes can affect
production payloads even when a diff looks small, so contributions must make
behavior and compatibility impact explicit.

## Before opening a change

- Search existing issues and pull requests.
- Open a design issue before adding a format, dependency, public abstraction,
  normalization rule, or backward-incompatible behavior.
- Redact fixtures. Never commit credentials, personal data, customer IDs, or
  proprietary payloads without permission.

## Development

Use Go 1.25.8 or newer.

Install the pinned CI tool families before running the complete local gate:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go install golang.org/x/vuln/cmd/govulncheck@latest
```

CI pins its scanner action. Contributors should refresh `govulncheck` before a
release or security-sensitive change because vulnerability data and the scanner
evolve independently of this module.

```sh
git clone https://github.com/faustbrian/go-wire.git
cd go-wire
make check
```

Focused targets include `make test`, `make coverage`, `make fuzz`,
`make benchmark`, `make docs`, `make lint`, and `make vuln`.

The runtime code should prefer the standard library. A dependency proposal
must explain its maintenance benefit, supported Go/platform constraints,
security posture, and why a small local implementation is not safer.

## Tests

Behavior changes require a red-green regression test. Keep 100% production
statement coverage, but prioritize assertions that prove:

- successful values and exact emitted bytes;
- malformed input and error classification;
- size, target, and option failures;
- namespace and charset behavior;
- SOAP envelope/fault semantics;
- discovered vendor regressions.

Add a redacted fixture when the shape came from a real interoperability issue.
Update fuzz seeds for a new parser family and benchmarks for a representative
new parse or encode path.

## Documentation

Update `CHANGELOG.md` for every file-changing contribution. Update the README,
API reference, matrices, examples, cookbook, FAQ, troubleshooting, and
migration notes when their public contract changes.

Documentation must state intentional divergences and limits. Do not claim broad
support from a single fixture or hide normalization behavior.

## Commit and pull request expectations

- Keep commits focused and use Conventional Commits with explanatory bodies.
- Do not combine format behavior with unrelated refactors.
- Describe why the behavior is required, compatibility impact, fixtures added,
  and exact verification commands.
- Complete the pull request template and wait for all required checks.

By contributing, you agree that your contribution is licensed under the MIT
License and that you will follow [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
