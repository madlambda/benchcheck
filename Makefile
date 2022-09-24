version?=$(shell git rev-list -1 HEAD)
cov=coverage.out
covhtml=coverage.html
golangci_lint_version=v1.49.0

COVERAGE_REPORT ?= coverage.txt

all: lint test build

.PHONY: build
build:
	go build -o ./cmd/benchcheck/benchcheck ./cmd/benchcheck

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_lint_version) run ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: test/integration
test/integration:
	go test -race -tags integration -count=1 ./...

.PHONY: coverage
coverage: 
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_REPORT) -tags integration ./...

.PHONY: coverage/show
coverage/show: coverage
	go tool cover -html=$(COVERAGE_REPORT)

.PHONY: install
install:
	go install ./cmd/benchcheck

.PHONY: cleanup
cleanup:
	rm -f cmd/benchcheck/benchcheck
