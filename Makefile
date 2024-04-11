.PHONY: build install test help
default: help

build: generate test
	go build -o bin/goenums goenums.go

install:
	chmod +x bin/goenums
	cp bin/goenums /usr/local/go/bin/goenums
	
test:
	go test -v ./...

generate:
	go generate ./...

help:
	@echo "build - build the goenums binary"
	@echo "install - install the goenums binary to /usr/local/go/bin *root/sudo required"
	@echo "test - run tests"
	@echo "help - print this help message"