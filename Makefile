GOPATH=$(shell go env GOPATH)
PROTOC_ZIP=protoc-3.14.0-osx-x86_64.zip

install-protoc:
	curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/$(PROTOC_ZIP)
	sudo unzip -o $(PROTOC_ZIP) -d /usr/local bin/protoc
	sudo unzip -o $(PROTOC_ZIP) -d /usr/local 'include/*'
	rm -f $(PROTOC_ZIP)

setup:
	go get google.golang.org/grpc
	go get github.com/golang/protobuf/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

proto:
	PATH=/usr/bin:/usr/local/bin:$(GOPATH)/bin protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative registry.proto


build: proto
	go build -o builds/rpc-registry cmd/server/rpc


test:
	go test -timeout 30s -cover -race ./...