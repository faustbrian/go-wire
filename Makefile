GO ?= go
FUZZ_TIME ?= 2s
BENCH_TIME ?= 1x

.PHONY: benchmark check coverage docs format format-check fuzz lint test vet vuln

format:
	gofmt -w .

format-check:
	test -z "$$(gofmt -l .)"

test:
	$(GO) test -race ./...

coverage:
	./scripts/check-coverage.sh

vet:
	$(GO) vet ./...

lint:
	golangci-lint run --timeout=5m

fuzz:
	$(GO) test -run '^$$' -fuzz=FuzzDecode -fuzztime=$(FUZZ_TIME) ./jsonwire
	$(GO) test -run '^$$' -fuzz=FuzzDecode -fuzztime=$(FUZZ_TIME) ./xmlwire
	$(GO) test -run '^$$' -fuzz=FuzzParse -fuzztime=$(FUZZ_TIME) ./soap

benchmark:
	$(GO) test -run '^$$' -bench . -benchtime=$(BENCH_TIME) ./...

docs:
	./scripts/check-docs.sh
	$(GO) test ./...

vuln:
	govulncheck ./...

check: format-check vet lint test coverage fuzz benchmark docs vuln
