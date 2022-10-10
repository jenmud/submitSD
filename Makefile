ADDR=localhost:8081
#MODE=debug
MODE=release

all: build

generate:
	go generate ./...

build-linux:
	go build -o builds/submitSD ./main

build-windows:
	go build -o builds/submitSD.exe ./main

build: generate build-linux build-windows

run:
	GIN_MODE=$(MODE) go run main/main.go server --addr $(ADDR)
