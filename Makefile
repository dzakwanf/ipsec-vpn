# Makefile for IPsec VPN with Post-Quantum Encryption

# Variables
BINARY_NAME=ipsec-vpn
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/dzakwan/ipsec-vpn/cmd.Version=$(VERSION) -X github.com/dzakwan/ipsec-vpn/cmd.Commit=$(COMMIT) -X github.com/dzakwan/ipsec-vpn/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: all build clean test install uninstall fmt lint vet

all: build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Install the binary
install: build
	mkdir -p /usr/local/bin
	cp $(BINARY_NAME) /usr/local/bin/

# Uninstall the binary
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golint ./...

# Run vet
vet:
	go vet ./...

# Run all code quality checks
check: fmt lint vet

# Build for multiple platforms
build-all: clean
	# Linux (amd64)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	# Linux (arm64)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .

# Create a release package
release: build-all
	mkdir -p release
	cp $(BINARY_NAME)-linux-amd64 release/
	cp $(BINARY_NAME)-linux-arm64 release/
	cp README.md release/
	cp .ipsec-vpn.yaml release/ipsec-vpn.yaml.example
	tar -czf $(BINARY_NAME)-$(VERSION).tar.gz -C release .
	rm -rf release

# Show help
help:
	@echo "Available targets:"
	@echo "  all        : Build the binary (default)"
	@echo "  build      : Build the binary"
	@echo "  clean      : Remove build artifacts"
	@echo "  test       : Run tests"
	@echo "  install    : Install the binary to /usr/local/bin"
	@echo "  uninstall  : Remove the binary from /usr/local/bin"
	@echo "  fmt        : Format code"
	@echo "  lint       : Run linter"
	@echo "  vet        : Run vet"
	@echo "  check      : Run all code quality checks"
	@echo "  build-all  : Build for multiple platforms"
	@echo "  release    : Create a release package"
	@echo "  help       : Show this help message"