SHELL := /bin/sh

BINARY_NAME := subspace-assignment

.PHONY: tidy fmt lint test run ci

tidy:
	go mod tidy

fmt:
	gofmt -w .

lint:
	golangci-lint run

test:
	go test ./...

run:
	go run ./cmd/$(BINARY_NAME) --help

ci: fmt lint test
