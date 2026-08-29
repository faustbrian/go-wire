GO ?= go
FUZZ_TIME ?= 2s
BENCH_TIME ?= 100ms

.PHONY: benchmark conformance docs fuzz

benchmark:
	$(GO) test -run '^$$' -bench . -benchmem -benchtime="$(BENCH_TIME)" ./...

conformance:
	./scripts/check-conformance.sh

docs:
	./scripts/check-docs.sh

fuzz:
	./scripts/check-fuzz.sh "$(FUZZ_TIME)"
