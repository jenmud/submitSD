proto:
	protoc --go_out=./registry/proto --go_opt=paths=source_relative \
	--go-grpc_out=./registry/proto --go-grpc_opt=paths=source_relative \
	./registry/proto/registry.proto