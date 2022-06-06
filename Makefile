version?=$(shell git rev-list -1 HEAD)
cov=coverage.out
covhtml=coverage.html
buildflags=-ldflags "-X main.Version=${version}"
golangci_lint_version=v1.45.0
name=benchcheck

COVERAGE_REPORT ?= coverage.txt

all: lint test build

.PHONY: build
build:
	go build $(buildflags) -o ./cmd/$(name)/$(name) ./cmd/$(name)

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_lint_version) run ./...

.PHONY: test
test:
	go test -timeout 200s -race ./...

.PHONY: coverage
coverage: 
	go test -covermode=atomic -coverprofile=$(COVERAGE_REPORT) ./...

.PHONY: coverage/show
coverage/show: coverage
	go tool cover -html=$(COVERAGE_REPORT)

.PHONY: install
install:
	go install $(buildflags) ./cmd/$(name)

.PHONY: cleanup
cleanup:
	rm -f cmd/$(name)/$(name)
