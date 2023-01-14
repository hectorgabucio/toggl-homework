CUR_DIR = $(CURDIR)
all: check-style test

## Runs golangci-lint
.PHONY: check-style
check-style:
	golangci-lint run ./...

## Builds project
.PHONY: build
build:
	go build cmd/main.go

## Runs tests
.PHONY: test
test:
	go test ./...

fmt:
	go fmt ./...