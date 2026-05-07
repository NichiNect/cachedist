.PHONY: run-node1 build test bench

build:
	go build -o bin/server ./cmd/server

run-node1:
	CACHE_NODE_ID=node-1 CACHE_HTTP_PORT=7001 CACHE_GRPC_PORT=8001 go run ./cmd/server/main.go

run-node2:
	CACHE_NODE_ID=node-2 CACHE_HTTP_PORT=7002 CACHE_GRPC_PORT=8002 go run ./cmd/server/main.go

run-node3:
	CACHE_NODE_ID=node-3 CACHE_HTTP_PORT=7003 CACHE_GRPC_PORT=8003 go run ./cmd/server/main.go

test:
	go test -v ./...

bench:
	go test -bench=. ./... --benchmem
