ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run test fmt

run:
	go run ./cmd

test:
	go test ./...

fmt:
	gofmt -w .
