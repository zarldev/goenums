# Build variables
VERSION := v0.4.3
BUILD_TIME := $(shell date +%Y%m%d-%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY := $(shell if [ -n "$$(git status --porcelain)" ]; then echo "-dirty"; fi)

# Properly formatted LDFLAGS
LDFLAGS := -ldflags "-X github.com/zarldev/goenums/internal/version.CURRENT='$(VERSION)' -X github.com/zarldev/goenums/internal/version.BUILD='$(BUILD_TIME)' -X github.com/zarldev/goenums/internal/version.COMMIT='$(GIT_COMMIT)$(GIT_DIRTY)'"
PRODLDFLAGS := -ldflags "-s -w -X github.com/zarldev/goenums/internal/version.CURRENT='$(VERSION)' -X github.com/zarldev/goenums/internal/version.BUILD='$(BUILD_TIME)' -X github.com/zarldev/goenums/internal/version.COMMIT='$(GIT_COMMIT)$(GIT_DIRTY)'"

# Fuzz test names
FUZZ_TESTS := FuzzParseValue_String FuzzParseValue_Int FuzzParseValue_Bool FuzzParseValue_Float64 FuzzParseValue_Duration FuzzParseEnumAliases FuzzParseEnumFields FuzzExtractFields

# Default target - what happens when you just run 'make'
.DEFAULT_GOAL := build

# Phony targets to avoid conflicts with files of the same name
.PHONY: build build-prod build-linux build-darwin build-windows deps test test-coverage test-fuzz test-fuzz-quick test-fuzz-long generate clean install uninstall lint help version logo debug-version release-tag release-tag-force release-build release-all

release-tag:
	@echo "🔍 Checking for uncommitted changes..."
	@if [ "$$(git status --porcelain | wc -l)" -ne "0" ]; then \
		echo "❌ Error: Working directory has uncommitted changes. Commit or stash them first."; \
		exit 1; \
	fi
	@echo "🏷️  Creating git tag $(VERSION)..."
	@if git tag -l | grep -q "^$(VERSION)$$"; then \
		echo "⚠️  Tag $(VERSION) already exists!"; \
		echo "💡 Options:"; \
		echo "   1. Delete existing tag: git tag -d $(VERSION)"; \
		echo "   2. Update VERSION in Makefile to a new version"; \
		echo "   3. Force recreate tag: make release-tag-force"; \
		exit 1; \
	fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "✅ Tag created. Push with: git push origin $(VERSION)"

# Force recreate an existing tag (deletes and recreates)
release-tag-force:
	@echo "🔍 Checking for uncommitted changes..."
	@if [ "$$(git status --porcelain | wc -l)" -ne "0" ]; then \
		echo "❌ Error: Working directory has uncommitted changes. Commit or stash them first."; \
		exit 1; \
	fi
	@echo "🗑️  Deleting existing tag $(VERSION) if it exists..."
	@git tag -d $(VERSION) 2>/dev/null || true
	@echo "🏷️  Creating git tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "✅ Tag created. Push with: git push origin $(VERSION) --force"

# Release build that ensures a clean state
release-build: build-prod
	@echo "🚀 Built release version $(VERSION)"

# Build all platforms from clean tagged state
release-all: build-linux-archive build-darwin-archive build-windows-archive
	@echo "🌍 Built all platforms and archived for release $(VERSION)"

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
	@echo "🐛 Debug Information:"
	@echo "   VERSION: $(VERSION)"
	@echo "   BUILD_TIME: $(BUILD_TIME)"
	@echo "   GIT_COMMIT: $(GIT_COMMIT)"
	@echo "   GIT_DIRTY: $(GIT_DIRTY)"
	@echo "   LDFLAGS: $(LDFLAGS)"

# Build with clear output
build: deps test
	@echo "🔨 Building goenums..."
	mkdir -p bin
	go build  $(LDFLAGS) -o bin/goenums goenums.go
	@echo "✅ Build completed with version $(VERSION) ($(BUILD_TIME), $(GIT_COMMIT)$(GIT_DIRTY))"

deps:
	@echo "📦 Managing dependencies..."
	go mod tidy
	go mod verify
	@echo "✅ Dependencies updated"

# Production build command - explicitly uses the prod tag
build-prod:
	@echo "🏭 Building production binary..."
	go build -trimpath -tags=prod $(PRODLDFLAGS) -o bin/goenums goenums.go
	@echo "✅ Production build completed"

# Other platform-specific builds
build-linux: generate test
	@echo "🐧 Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -tags=prod $(LDFLAGS) -o bin/linux/amd64/goenums goenums.go
	GOOS=linux GOARCH=arm64 go build -tags=prod $(LDFLAGS) -o bin/linux/arm64/goenums goenums.go
	@echo "✅ Linux builds completed"

build-darwin: generate test
	@echo "🍎 Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -tags=prod $(LDFLAGS) -o bin/darwin/amd64/goenums goenums.go
	GOOS=darwin GOARCH=arm64 go build -tags=prod $(LDFLAGS) -o bin/darwin/arm64/goenums goenums.go
	@echo "✅ macOS builds completed"

build-windows: generate test
	@echo "🪟 Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -tags=prod $(LDFLAGS) -o bin/windows/amd64/goenums.exe goenums.go
	@echo "✅ Windows build completed"

install:
	@echo "📥 Installing goenums..."
	chmod +x bin/goenums
	@echo "   Installing to /usr/local/bin/goenums"
	@if [ -w /usr/local/bin ]; then \
		cp bin/goenums /usr/local/bin/goenums; \
		echo "✅ Installation completed"; \
	else \
		echo "🔐 Need sudo permission to install"; \
		sudo cp bin/goenums /usr/local/bin/goenums; \
		echo "✅ Installation completed"; \
	fi

uninstall:
	@echo "🗑️  Uninstalling goenums..."
	@if [ -f /usr/local/bin/goenums ]; then \
		if [ -w /usr/local/bin ]; then \
			rm /usr/local/bin/goenums; \
			echo "✅ Uninstallation completed"; \
		else \
			echo "🔐 Need sudo permission to uninstall"; \
			sudo rm /usr/local/bin/goenums; \
			echo "✅ Uninstallation completed"; \
		fi; \
	else \
		echo "ℹ️  goenums is not installed in /usr/local/bin"; \
	fi

test:
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"
	@cd internal/testdata && go test -v $(shell cd internal/testdata && go list ./... | grep -v notgocode)
	@echo "✅ Testdata Tests completed"

test-coverage:
	@echo "📊 Running tests with coverage..."
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	@echo "🔍 Filtering coverage profile to exclude examples..."
	@grep -v "github.com/zarldev/goenums/example" cover.out > cover_filtered.out 2>/dev/null || cp cover.out cover_filtered.out
	@mv cover_filtered.out cover.out
	go-test-coverage --config=./.testcoverage.yml
	@echo "📈 Generating HTML coverage report..."
	go tool cover -html=cover.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run all fuzz tests for 30 seconds each
test-fuzz:
	@echo "🧪 Running fuzz tests (30s each)..."
	@total=$$(echo "$(FUZZ_TESTS)" | wc -w); \
	current=1; \
	for test in $(FUZZ_TESTS); do \
		echo "[$${current}/$${total}] Running $${test}..."; \
		if go test -fuzz=$${test} -fuzztime=30s ./enum; then \
			echo "✅ $${test} completed successfully"; \
		else \
			echo "❌ $${test} failed"; \
			exit 1; \
		fi; \
		current=$$((current + 1)); \
		echo ""; \
	done; \
	echo "🎉 All fuzz tests completed successfully!"

# Run fuzz tests for a longer duration (useful for CI or thorough testing)
test-fuzz-long:
	@echo "🧪 Running extended fuzz tests (2m each)..."
	@total=$$(echo "$(FUZZ_TESTS)" | wc -w); \
	current=1; \
	for test in $(FUZZ_TESTS); do \
		echo "[$${current}/$${total}] Running $${test} for 2 minutes..."; \
		if go test -fuzz=$${test} -fuzztime=2m ./enum; then \
			echo "✅ $${test} completed successfully"; \
		else \
			echo "❌ $${test} failed"; \
			exit 1; \
		fi; \
		current=$$((current + 1)); \
		echo ""; \
	done; \
	echo "🎉 All extended fuzz tests completed successfully!"

# Quick fuzz test run (10s each) for development
test-fuzz-quick:
	@echo "🧪 Running quick fuzz tests (10s each)..."
	@total=$$(echo "$(FUZZ_TESTS)" | wc -w); \
	current=1; \
	for test in $(FUZZ_TESTS); do \
		echo "[$${current}/$${total}] Running $${test}..."; \
		if go test -fuzz=$${test} -fuzztime=10s ./enum; then \
			echo "✅ $${test} completed"; \
		else \
			echo "❌ $${test} failed"; \
			exit 1; \
		fi; \
		current=$$((current + 1)); \
	done; \
	echo "🎉 Quick fuzz tests completed!"

generate:
	@echo "⚙️  Running code generation..."
	go generate ./...
	@echo "✅ Code generation completed"

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	@go clean -testcache
	@echo "✅ Clean completed"

version: logo
	@echo "              version: $(VERSION)"
	@echo "              built:   $(BUILD_TIME)"
	@echo "              commit:  $(GIT_COMMIT)$(GIT_DIRTY)"

lint:
	@echo "🔍 Running linter..."
	golangci-lint run ./...
	@echo "✅ Linting completed"

logo:
	@echo "   ____ _____  ___  ____  __  ______ ___  _____"
	@echo "  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/"
	@echo " / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) "
	@echo " \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  "
	@echo "/____/ "

help:
	@echo "📚 Available commands:"
	@echo ""
	@echo "🔨 Build Commands:"
	@echo "  build             - build the goenums binary for current platform"
	@echo "  build-prod        - build production binary with optimizations"
	@echo "  build-linux       - build for Linux (amd64, arm64)"
	@echo "  build-darwin      - build for macOS (amd64, arm64)"
	@echo "  build-windows     - build for Windows (amd64)"
	@echo "  build-all         - build for all supported platforms"
	@echo ""
	@echo "🚀 Release Commands:"
	@echo "  release-tag       - create a git tag for release"
	@echo "  release-tag-force - force recreate an existing git tag"
	@echo "  release-build     - build release version"
	@echo "  release-all       - build and archive all platforms"
	@echo ""
	@echo "🧪 Testing Commands:"
	@echo "  test              - run tests"
	@echo "  test-coverage     - run tests with coverage report"
	@echo "  test-fuzz         - run all fuzz tests for 30s each"
	@echo "  test-fuzz-quick   - run all fuzz tests for 10s each (development)"
	@echo "  test-fuzz-long    - run all fuzz tests for 2m each (thorough)"
	@echo ""
	@echo "🛠️  Development Commands:"
	@echo "  deps              - manage dependencies"
	@echo "  generate          - run go generate"
	@echo "  lint              - run linter"
	@echo "  clean             - remove build artifacts"
	@echo "  debug-version     - show build variables"
	@echo ""
	@echo "📦 Installation:"
	@echo "  install           - install the goenums binary to /usr/local/bin"
	@echo ""
	@echo "ℹ️  Information:"
	@echo "  help              - print this help message"
	@echo "  version           - print the version"