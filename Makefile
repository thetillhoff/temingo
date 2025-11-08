.PHONY: list
list:
	@grep -E '^[[:alpha:]].*:' Makefile | cat # Get all targets in this file, without color-coding the matching letters

install:
	go get ./...

run:
	go run ./...

test:
	go test -v ./...

build:
	go build

format:
	go fmt ./...

upgrade:
	go get -u ./...
	go mod tidy
