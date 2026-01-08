# Memento Makefile

BINARY_NAME := memento
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

GO := go
GOFMT := gofmt
GOFLAGS :=

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man/man1
CONFIGDIR := $(HOME)/.config/memento

.PHONY: all build clean install uninstall fmt lint test run help init-config

all: build

build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/memento

build-debug:
	$(GO) build -gcflags="all=-N -l" -o $(BINARY_NAME) ./cmd/memento

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

install: build
	install -d $(BINDIR)
	install -m 755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)
	@if [ -f docs/memento.1 ]; then \
		install -d $(MANDIR); \
		install -m 644 docs/memento.1 $(MANDIR)/memento.1; \
	fi
	@echo "Installed to $(BINDIR)/$(BINARY_NAME)"

uninstall:
	rm -f $(BINDIR)/$(BINARY_NAME)
	rm -f $(MANDIR)/memento.1
	@echo "Uninstalled from $(BINDIR)"

fmt:
	$(GOFMT) -w -s .

lint:
	$(GO) vet ./...
	@if command -v staticcheck > /dev/null; then staticcheck ./...; fi
	@if command -v golangci-lint > /dev/null; then golangci-lint run; fi

test:
	$(GO) test -v ./...

test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

run: build
	./$(BINARY_NAME)

run-debug: build
	./$(BINARY_NAME) --debug

init-config: build
	./$(BINARY_NAME) --init-config

dist:
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/memento
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/memento
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/memento
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/memento
	@echo "Binaries built in dist/"

deps:
	$(GO) mod download
	$(GO) mod tidy

man:
	@if [ -f docs/memento.1.md ]; then \
		pandoc docs/memento.1.md -s -t man -o docs/memento.1; \
		echo "Man page generated: docs/memento.1"; \
	else \
		echo "No docs/memento.1.md found"; \
	fi

help:
	@echo "Memento Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the binary (default)"
	@echo "  build-debug  Build with debug symbols"
	@echo "  clean        Remove build artifacts"
	@echo "  install      Install binary and man page to $(PREFIX)"
	@echo "  uninstall    Remove installed files"
	@echo "  fmt          Format Go source files"
	@echo "  lint         Run linters"
	@echo "  test         Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  run          Build and run the application"
	@echo "  run-debug    Build and run with debug mode"
	@echo "  init-config  Create default config file"
	@echo "  dist         Cross-compile for multiple platforms"
	@echo "  deps         Download and tidy dependencies"
	@echo "  man          Generate man page (requires pandoc)"
	@echo "  help         Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX       Installation prefix (default: /usr/local)"
	@echo "  BINDIR       Binary installation directory (default: $(PREFIX)/bin)"
	@echo "  MANDIR       Man page installation directory (default: $(PREFIX)/share/man/man1)"
