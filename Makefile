all: build

generate:
	go generate ./...

build-linux:
	go build -o builds/serviceSD ./main

build-windows:
	go build -o builds/serviceSD.exe ./main

build: generate build-linux build-windows