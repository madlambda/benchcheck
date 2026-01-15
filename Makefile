version?=$(shell git rev-list -1 HEAD)
cov=coverage.out
covhtml=coverage.html
buildflags=-ldflags "-X main.Version=${version}"
golangci_lint_version=v1.49.0
coverage_report ?= coverage.txt


all: lint test build

.PHONY: build
build:
	go build $(buildflags) -o ./cmd/benchcheck/benchcheck ./cmd/benchcheck

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_lint_version) run ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: coverage
coverage: 
	go test -race -covermode=atomic -coverprofile=$(coverage_report) -tags integration ./...

.PHONY: coverage/show
coverage/show: coverage
	go tool cover -html=$(coverage_report)

.PHONY: install
install:
	go install $(buildflags) ./cmd/benchcheck

.PHONY: cleanup
cleanup:
	rm -f cmd/benchcheck/benchcheck
