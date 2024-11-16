.PHONY: setup test lint security-check build clean

setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	pip install pre-commit
	pre-commit install

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

build:
	go build -v ./...

clean:
	go clean
	rm -f coverage.out

all: setup lint test build
