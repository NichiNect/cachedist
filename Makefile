.PHONY: run-node1 build test bench

build:
	go build -o bin/server ./cmd/server

run-node1:
	CACHE_NODE_ID=node-1 CACHE_HTTP_PORT=7001 go run ./cmd/server/main.go

test:
	go test -v ./...

bench:
	go test -bench=. ./...
