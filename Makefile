.PHONY: build install clean test lint format vet tidy docs

BINARY_NAME := flat
VERSION := 0.1.0

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION)"

build-debug:
	@echo "Building $(BINARY_NAME) (debug)..."
	go build -o $(BINARY_NAME) -gcflags "all=-N -l"

build-release:
	@echo "Building $(BINARY_NAME) (release)..."
	go build -o $(BINARY_NAME) -ldflags "-s -w -X main.version=$(VERSION)"

install: build
	@echo "Installing $(BINARY_NAME) to \$GOPATH/bin..."
	go install -ldflags "-X main.version=$(VERSION)"
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo "Uninstalled"

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Cleaned"

test:
	@echo "Running tests..."
	go test ./... -v -cover

test-short:
	@echo "Running tests (short)..."
	go test ./... -cover

test-race:
	@echo "Running tests with race detector..."
	go test ./... -race -v

coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -covermode=count
	@echo "Coverage report generated (coverage.out)"

coverage-html: coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o=coverage.html
	@echo "HTML coverage report generated (coverage.html)"

lint:
	@echo "Running linters..."
	go vet ./...
	@gofmt -l .
	@echo "Linting passed"

format:
	@echo "Formatting code..."
	gofmt -w .
	@echo "Code formatted"

vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Go vet passed"

tidy:
	@echo "Running go mod tidy..."
	go mod tidy
	@echo "Dependencies tidied"

cross-build:
	@echo "Building for multiple platforms..."
	@for platform in "linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64"; do \
		IFS='/' read -r goos goarch <<< "$$platform"; \
		out="$(BINARY_NAME)-$$goos-$$goarch"; \
		if [ "$$goos" = "windows" ]; then out="$$out.exe"; fi; \
		echo "Building for $$goos/$$goarch..."; \
		GOOS=$$goos GOARCH=$$goarch go build -o "$$out" -ldflags "-X main.version=$(VERSION)"; \
		echo "✓ Built $$out"; \
	done
	@echo "Cross-build complete"

docs:
	@echo "Documentation:"
	@echo "  - docs/phase-0.md"
	@echo "  - docs/phase-1.md"
	@echo "  - docs/phase-2.md"
	@echo "  - docs/phase-3.md"
	@echo "  - docs/overview.md"
	@echo "  - docs/checklist.md"
	@echo "  - docs/test-coverage.md"
	@echo "  - README.md"

help:
	@echo "Flat Build Commands"
	@echo "==================="
	@echo ""
	@echo "Build:"
	@echo "  make build          Build flat binary"
	@echo "  make build-debug    Build with debug symbols"
	@echo "  make build-release  Build optimized release binary"
	@echo "  make cross-build    Build for multiple platforms"
	@echo ""
	@echo "Install:"
	@echo "  make install        Install to \$GOPATH/bin"
	@echo "  make uninstall      Uninstall from \$GOPATH/bin"
	@echo ""
	@echo "Test:"
	@echo "  make test           Run all tests"
	@echo "  make test-short     Run tests without verbose output"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make coverage       Run tests with coverage report"
	@echo "  make coverage-html  Generate HTML coverage report"
	@echo ""
	@echo "Lint:"
	@echo "  make lint           Run linters (go vet, gofmt)"
	@echo "  make format         Format code with gofmt"
	@echo "  make vet            Run go vet"
	@echo ""
	@echo "Clean:"
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Other:"
	@echo "  make tidy           Run go mod tidy"
	@echo "  make docs           Display documentation info"
	@echo "  make help           Display this help"

