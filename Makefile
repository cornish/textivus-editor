.PHONY: build test fmt lint setup clean

# Build the binary
build:
	go build -o textivus

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	gofmt -w .

# Run linters
lint:
	go vet ./...
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code needs formatting. Run 'make fmt'"; \
		gofmt -l .; \
		exit 1; \
	fi

# Setup development environment (run once after clone)
setup:
	git config core.hooksPath .githooks
	@echo "Git hooks configured. Pre-commit hook will auto-format Go files."

# Clean build artifacts
clean:
	rm -f textivus
	rm -f textivus-*
