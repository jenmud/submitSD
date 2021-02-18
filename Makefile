setup:
	go get google.golang.org/grpc
	go get github.com/golang/protobuf/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative registry.proto


build: proto
	go build -o builds/rpc-registry cmd/server/rpc