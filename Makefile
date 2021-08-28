version?=$(shell git rev-list -1 HEAD)
cov=coverage.out
covhtml=coverage.html
buildflags=-ldflags "-X main.Version=${version}"
golangci_lint_version=1.41.1
name=benchcheck

all: lint test build

.PHONY: build
build:
	go build $(buildflags) -o ./cmd/$(name)/$(name) ./cmd/$(name)

.PHONY: lint
lint:
	docker run --rm -v `pwd`:/app -w /app golangci/golangci-lint:v$(golangci_lint_version)  golangci-lint run ./...

.PHONY: test
test:
	go test -timeout 10s -race -coverprofile=$(cov) ./...

.PHONY: coverage
coverage: test
	go tool cover -html=$(cov) -o=$(covhtml)

.PHONY: install
install:
	go install $(buildflags) ./cmd/$(name)

.PHONY: cleanup
cleanup:
	rm -f cmd/$(name)/$(name)
