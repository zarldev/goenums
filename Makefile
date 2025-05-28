# Build variables
VERSION := v0.4.0
BUILD_TIME := $(shell date +%Y%m%d-%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY := $(shell if [ -n "$$(git status --porcelain)" ]; then echo "-dirty"; fi)

# Properly formatted LDFLAGS
LDFLAGS := -ldflags "-X github.com/zarldev/goenums/internal/version.CURRENT='$(VERSION)' -X github.com/zarldev/goenums/internal/version.BUILD='$(BUILD_TIME)' -X github.com/zarldev/goenums/internal/version.COMMIT='$(GIT_COMMIT)$(GIT_DIRTY)'"
PRODLDFLAGS := -ldflags "-s -w -X github.com/zarldev/goenums/internal/version.CURRENT='$(VERSION)' -X github.com/zarldev/goenums/internal/version.BUILD='$(BUILD_TIME)' -X github.com/zarldev/goenums/internal/version.COMMIT='$(GIT_COMMIT)$(GIT_DIRTY)'"

release-tag:
	@echo "Checking for uncommitted changes..."
	@if [ "$$(git status --porcelain | wc -l)" -ne "0" ]; then \
		echo "Error: Working directory has uncommitted changes. Commit or stash them first."; \
		exit 1; \
	fi
	@echo "Creating git tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag created. Push with: git push origin $(VERSION)"

# Release build that ensures a clean state
release-build: build-prod
	@echo "Built release version $(VERSION)"

# Build all platforms from clean tagged state
release-all: build-linux-archive build-darwin-archive build-windows-archive
	@echo "Built all platforms and archived for release $(VERSION)"

build-linux-archive: build-linux
	mkdir -p dist
	cp bin/linux/amd64/goenums dist/goenums-$(VERSION)-linux-amd64
	cp bin/linux/arm64/goenums dist/goenums-$(VERSION)-linux-arm64
	tar -czf dist/goenums-$(VERSION)-linux-amd64.tar.gz -C dist goenums-$(VERSION)-linux-amd64
	tar -czf dist/goenums-$(VERSION)-linux-arm64.tar.gz -C dist goenums-$(VERSION)-linux-arm64
	rm dist/goenums-$(VERSION)-linux-amd64
	rm dist/goenums-$(VERSION)-linux-arm64

build-darwin-archive: build-darwin
	mkdir -p dist
	cp bin/darwin/amd64/goenums dist/goenums-$(VERSION)-darwin-amd64
	cp bin/darwin/arm64/goenums dist/goenums-$(VERSION)-darwin-arm64
	tar -czf dist/goenums-$(VERSION)-darwin-amd64.tar.gz -C dist goenums-$(VERSION)-darwin-amd64
	tar -czf dist/goenums-$(VERSION)-darwin-arm64.tar.gz -C dist goenums-$(VERSION)-darwin-arm64
	rm dist/goenums-$(VERSION)-darwin-amd64
	rm dist/goenums-$(VERSION)-darwin-arm64

build-windows-archive: build-windows
	mkdir -p dist
	cp bin/windows/amd64/goenums.exe dist/goenums-$(VERSION)-windows-amd64.exe
	tar -czf dist/goenums-$(VERSION)-windows-amd64.tar.gz -C dist goenums-$(VERSION)-windows-amd64.exe
	rm dist/goenums-$(VERSION)-windows-amd64.exe

# Debug target to verify variable values
debug-version:
	@echo "VERSION: $(VERSION)"
	@echo "BUILD_TIME: $(BUILD_TIME)"
	@echo "GIT_COMMIT: $(GIT_COMMIT)"
	@echo "GIT_DIRTY: $(GIT_DIRTY)"
	@echo "LDFLAGS: $(LDFLAGS)"

# Build with clear output
build: deps test
	mkdir -p bin
	go build  $(LDFLAGS) -o bin/goenums goenums.go
	@echo "Build with version $(VERSION) ($(BUILD_TIME), $(GIT_COMMIT)$(GIT_DIRTY))"

deps:
	go mod tidy
	go mod verify

# Production build command - explicitly uses the prod tag
build-prod:
	go build -trimpath -tags=prod $(PRODLDFLAGS) -o bin/goenums goenums.go

# Other platform-specific builds
build-linux: generate test
	GOOS=linux GOARCH=amd64 go build -tags=prod $(LDFLAGS) -o bin/linux/amd64/goenums goenums.go
	GOOS=linux GOARCH=arm64 go build -tags=prod $(LDFLAGS) -o bin/linux/arm64/goenums goenums.go

build-darwin: generate test
	GOOS=darwin GOARCH=amd64 go build -tags=prod $(LDFLAGS) -o bin/darwin/amd64/goenums goenums.go
	GOOS=darwin GOARCH=arm64 go build -tags=prod $(LDFLAGS) -o bin/darwin/arm64/goenums goenums.go

build-windows: generate test
	GOOS=windows GOARCH=amd64 go build -tags=prod $(LDFLAGS) -o bin/windows/amd64/goenums.exe goenums.go

install:
	chmod +x bin/goenums
	@echo "Installing to /usr/local/bin/goenums"
	@if [ -w /usr/local/bin ]; then \
		cp bin/goenums /usr/local/bin/goenums; \
	else \
		echo "Need sudo permission to install"; \
		sudo cp bin/goenums /usr/local/bin/goenums; \
	fi

test:
	go test -v ./...

test-coverage:
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	go-test-coverage --config=./.testcoverage.yml

generate:
	go generate ./...

clean:
	rm -rf bin/

version: logo
	@echo "              version: $(VERSION)"
	@echo "              built:   $(BUILD_TIME)"
	@echo "              commit:  $(GIT_COMMIT)$(GIT_DIRTY)"

lint:
	golangci-lint run ./...


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