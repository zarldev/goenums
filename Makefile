.PHONY: build install test help clean build-all build-linux build-darwin build-windows

# Build variables
VERSION := $(shell grep -o '".*"' internal/version/version.go | tr -d '"')
LDFLAGS := -ldflags "-X github.com/zarldev/goenums/pkg/version.CURRENT=$(VERSION)"
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

default: help

build: generate test
	go build $(LDFLAGS) -o bin/goenums goenums.go

# Build for all platforms
build-all: generate test $(PLATFORMS)

# Pattern rule for platform-specific builds
$(PLATFORMS):
	@echo "Building for $@"
	@mkdir -p bin/$@
	GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) go build $(LDFLAGS) -o bin/$@/goenums$(if $(findstring windows,$(word 1,$(subst /, ,$@))),.exe,) goenums.go

# Convenience targets for specific platforms
build-linux: generate test linux/amd64 linux/arm64

build-darwin: generate test darwin/amd64 darwin/arm64

build-windows: generate test windows/amd64

install:
	chmod +x bin/goenums
	cp bin/goenums /usr/local/go/bin/goenums
	
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out
	
generate:
	go generate ./...

clean:
	rm -rf bin/

version: logo
	@echo "              version: $(VERSION)"


logo:
	@echo "   ____ _____  ___  ____  __  ______ ___  _____"
	@echo "  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/"
	@echo " / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) "
	@echo " \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  "
	@echo "/____/ "


help:
	@echo "build       - build the goenums binary for current platform"
	@echo "build-all   - build for all supported platforms"
	@echo "build-linux - build for Linux (amd64, arm64)"
	@echo "build-darwin - build for macOS (amd64, arm64)"
	@echo "build-windows - build for Windows (amd64)"
	@echo "install     - install the goenums binary to /usr/local/go/bin *root/sudo required"
	@echo "test        - run tests"
	@echo "test-coverage - run tests with coverage report"
	@echo "generate    - run go generate"
	@echo "clean       - remove build artifacts"
	@echo "help        - print this help message"
	@echo "version     - print the version"