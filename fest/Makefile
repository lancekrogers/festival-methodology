.PHONY: all build test clean install uninstall run-help

# Binary name
BINARY_NAME=fest
MAIN_PATH=cmd/fest/main.go

# Build variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build flags
LDFLAGS=-ldflags "-s -w"

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -cover ./...

test-coverage-html:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

deps:
	$(GOGET) -u ./...
	$(GOMOD) tidy

install: build
	@echo "Installing fest..."
	@GOBIN=$$(go env GOPATH)/bin; \
	if [ -z "$$GOBIN" ]; then GOBIN=~/go/bin; fi; \
	mkdir -p $$GOBIN; \
	cp $(BINARY_NAME) $$GOBIN/$(BINARY_NAME); \
	if [[ "$$(uname)" == "Darwin" ]]; then \
		echo "Signing fest binary for macOS..."; \
		codesign --force --sign - $$GOBIN/$(BINARY_NAME) 2>/dev/null || \
		echo "Warning: Could not sign binary (non-fatal)"; \
	fi; \
	echo "✓ fest installed to $$GOBIN/$(BINARY_NAME)"

uninstall:
	@echo "Uninstalling fest..."
	@GOBIN=$$(go env GOPATH)/bin; \
	if [ -z "$$GOBIN" ]; then GOBIN=~/go/bin; fi; \
	if [ -f $$GOBIN/$(BINARY_NAME) ]; then \
		rm $$GOBIN/$(BINARY_NAME); \
		echo "✓ fest uninstalled from $$GOBIN"; \
	else \
		echo "fest not found in $$GOBIN"; \
	fi

run-help: build
	./$(BINARY_NAME) --help

run-init: build
	./$(BINARY_NAME) init

run-sync: build
	./$(BINARY_NAME) sync

run-update: build
	./$(BINARY_NAME) update

# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-all: build-linux build-darwin build-darwin-arm64 build-windows

# Code signing for release binaries
sign-release: build-darwin build-darwin-arm64
	@echo "Signing macOS release binaries..."
	@./scripts/sign-macos.sh $(BINARY_NAME)-darwin-amd64
	@./scripts/sign-macos.sh $(BINARY_NAME)-darwin-arm64

# Release targets
release-darwin: build-darwin sign-release
	@echo "Creating macOS release package..."
	@mkdir -p dist
	@tar -czf dist/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@tar -czf dist/$(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@echo "Release packages created in dist/"

release-linux: build-linux
	@echo "Creating Linux release package..."
	@mkdir -p dist
	@tar -czf dist/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64

release-windows: build-windows
	@echo "Creating Windows release package..."
	@mkdir -p dist
	@zip -q dist/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

release: release-darwin release-linux release-windows
	@echo "All release packages created in dist/"
	@ls -la dist/

# Development
dev:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

lint: fmt vet
	@echo "Linting complete"