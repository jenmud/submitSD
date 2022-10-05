all: build

generate:
	go generate ./...

build-linux:
	go build -o builds/submitSD ./main

build-windows:
	go build -o builds/submitSD.exe ./main

build: generate build-linux build-windows