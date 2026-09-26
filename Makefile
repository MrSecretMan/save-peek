.PHONY: test build run

test:
	go test ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o savepeek ./cmd/savepeek

run:
	go run ./cmd/savepeek
