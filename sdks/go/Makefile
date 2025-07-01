# RustPBX Go SDK Makefile

.PHONY: help build test clean examples build-examples run-basic run-webrtc run-ai run-sip

# Default target
help:
	@echo "RustPBX Go SDK - Available targets:"
	@echo "  build          - Build the SDK"
	@echo "  test           - Run tests"
	@echo "  examples       - Build all examples"
	@echo "  build-examples - Build all examples"
	@echo "  run-basic      - Run basic call example"
	@echo "  run-webrtc     - Run WebRTC demo"
	@echo "  run-ai         - Run AI voice assistant example"
	@echo "  run-sip        - Run SIP integration example"
	@echo "  clean          - Clean built binaries"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"

# Build the SDK
build:
	@echo "Building RustPBX Go SDK..."
	go build ./rustpbx

# Run tests
test:
	@echo "Running tests..."
	go test ./rustpbx -v

# Build all examples
examples: build-examples

build-examples: clean
	@echo "Building examples..."
	@mkdir -p bin
	go build -o bin/basic_call examples/basic_call.go
	go build -o bin/webrtc_demo examples/webrtc_demo.go
	go build -o bin/ai_voice_assistant examples/ai_voice_assistant.go
	go build -o bin/sip_integration examples/sip_integration.go
	@echo "Examples built in bin/ directory"

# Run individual examples
run-basic:
	@echo "Running basic call example..."
	go run examples/basic_call.go

run-webrtc:
	@echo "Running WebRTC demo..."
	go run examples/webrtc_demo.go

run-ai:
	@echo "Running AI voice assistant..."
	go run examples/ai_voice_assistant.go

run-sip:
	@echo "Running SIP integration example..."
	go run examples/sip_integration.go

# Run all examples (for testing)
run-all-examples:
	@echo "Running all examples with 10-second timeout each..."
	timeout 10s go run examples/basic_call.go || true
	timeout 10s go run examples/webrtc_demo.go || true
	timeout 10s go run examples/ai_voice_assistant.go || true
	timeout 10s go run examples/sip_integration.go || true

# Clean built binaries
clean:
	@echo "Cleaning built binaries..."
	rm -rf bin/
	rm -f basic_call webrtc_demo ai_voice_assistant sip_integration

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Full development setup
setup: deps install-tools
	@echo "Development environment setup complete"

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./rustpbx

# Check for security issues
security:
	@echo "Checking for security issues..."
	gosec ./...

# All checks (test, lint, security)
check: test lint security
	@echo "All checks passed!"

# CI target
ci: deps build test lint
	@echo "CI pipeline completed successfully"

# Release preparation
release-check: clean build test lint security
	@echo "Release check completed - ready for release"