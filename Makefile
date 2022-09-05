all: build

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	./registry/proto/registry.proto

build-linux:
	go build -o builds/serviceSD ./main

build-windows:
	go build -o builds/serviceSD.exe ./main

build: proto build-linux build-windows