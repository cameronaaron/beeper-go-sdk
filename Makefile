.PHONY: build test lint fmt vet clean install-tools generate

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOLINT=golangci-lint

# Build
build:
	$(GOBUILD) -v ./...

# Test
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Test with coverage report
test-coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint
lint:
	$(GOLINT) run

# Format
fmt:
	$(GOFMT) -s -w .

# Vet
vet:
	$(GOVET) ./...

# Clean
clean:
	$(GOCLEAN)
	rm -f coverage.out coverage.html

# Install development tools
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks
check: fmt vet lint test

# Generate Go code (if needed)
generate:
	$(GOCMD) generate ./...

# Update dependencies
update-deps:
	$(GOCMD) get -u ./...
	$(GOCMD) mod tidy

# Run example
run-example:
	$(GOCMD) run examples/basic/main.go

# Build test app
build-testapp:
	$(GOBUILD) -o testapp cmd/testapp/main.go

# Run test app
run-testapp: build-testapp
	./testapp

# Run test app in demo mode
demo:
	@echo "Running test app in demo mode..."
	@BEEPER_ACCESS_TOKEN="" ./testapp 2>/dev/null || $(GOCMD) run cmd/testapp/main.go
