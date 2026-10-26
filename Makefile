.PHONY: generate-grpc test-grpc

generate-grpc:
	./scripts/generate-grpc.sh

test-grpc:
	go test -race ./internal/rpc/...
