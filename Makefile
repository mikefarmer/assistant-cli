.PHONY: build test clean lint fmt help install build-all

# Variables
BINARY_NAME=assistant-cli
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Default target
all: test build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## build: Build the binary for the current platform
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} main.go

## build-all: Build binaries for macOS platforms
build-all: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o dist/${BINARY_NAME}-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o dist/${BINARY_NAME}-darwin-arm64 main.go

## install: Install the binary to GOPATH/bin
install:
	go install ${LDFLAGS}

## test: Run all tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Please install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

## fmt: Format all Go files
fmt:
	go fmt ./...

## clean: Clean build artifacts
clean:
	rm -f ${BINARY_NAME}
	rm -rf dist/
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## dev: Run the CLI in development mode
dev:
	go run main.go

## verify: Run fmt, lint, and test
verify: fmt lint test

## checksums: Generate checksums for built binaries
checksums:
	@if [ ! -d "dist" ]; then echo "No dist directory found. Run 'make build-all' first."; exit 1; fi
	cd dist && sha256sum * > checksums.txt
	@echo "Checksums generated in dist/checksums.txt"

## release-prep: Prepare for release (build-all + checksums)
release-prep: build-all checksums
	@echo "Release preparation complete. Files in dist/ directory:"
	@ls -la dist/

## release-preview: Show what a release would do without creating it
release-preview:
	@echo "🔍 Release Preview"
	@echo "=================="
	@echo "Current tags:"
	@git tag -l | tail -5
	@echo ""
	@read -p "Enter version to preview (e.g., v1.0.0): " version; \
	if [ -z "$$version" ]; then echo "Version cannot be empty"; exit 1; fi; \
	echo ""; \
	echo "📋 This would:"; \
	echo "  1. Build binaries with version: $$version"; \
	echo "  2. Test binary version matches: $$version"; \
	echo "  3. Create git tag: $$version"; \
	echo "  4. Push tag to trigger GitHub Actions release"; \
	echo "  5. GitHub will build and publish: assistant-cli-darwin-amd64, assistant-cli-darwin-arm64"; \
	echo ""; \
	echo "💡 To actually create the release, run: make tag-release"

## tag-release: Create and push a semantic version tag with version sync
tag-release:
	@echo "Current tags:"
	@git tag -l | tail -5
	@echo ""
	@read -p "Enter new version (e.g., v1.0.0): " version; \
	if [ -z "$$version" ]; then echo "Version cannot be empty"; exit 1; fi; \
	echo "Building release binaries with version $$version..."; \
	VERSION=$$version $(MAKE) build-all; \
	echo "Testing binary with version $$version..."; \
	ARCH=$$(uname -m); \
	if [ "$$ARCH" = "x86_64" ]; then ARCH="amd64"; elif [ "$$ARCH" = "aarch64" ]; then ARCH="arm64"; fi; \
	./dist/${BINARY_NAME}-darwin-$$ARCH --version | grep "$$version" || (echo "Version mismatch in binary!" && exit 1); \
	echo "Creating git tag $$version..."; \
	git tag -a "$$version" -m "Release $$version"; \
	echo "Pushing tag $$version to trigger release workflow..."; \
	git push origin "$$version"; \
	echo "✅ Release $$version created! Check GitHub Actions for build progress."