.PHONY: test build run

test:
	go test ./...

build:
	go build -trimpath -o savepeek ./cmd/savepeek

run:
	go run ./cmd/savepeek
