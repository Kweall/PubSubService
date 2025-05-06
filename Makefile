PROTO_DIR=proto
GEN_DIR=internal/delivery/grpc/gen

proto:
	protoc \
		--go_out=$(GEN_DIR) \
		--go-grpc_out=$(GEN_DIR) \
		--proto_path=$(PROTO_DIR) \
		$(PROTO_DIR)/pubsub.proto

build:
	go build -o bin/pubsub-server ./cmd/server

run:
	GRPC_PORT=50051 go run ./cmd/server

clean:
	rm -rf bin/
