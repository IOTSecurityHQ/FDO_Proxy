# FDO Server Proxy Makefile

.PHONY: build test clean help setup setup-fdo-backend docker-build docker-run docker-stop compose-up compose-down compose-logs auto-setup

# Default target
all: build

# Build the proxy
build:
	@echo "Building FDO Server Proxy..."
	go build -o fdo-proxy ./cmd/server
	@echo "Build complete: fdo-proxy"

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover ./...

# Run tests and generate coverage report
test-coverage-report:
	@echo "Running tests with coverage report..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f fdo-proxy
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatting complete"

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Setup FDO backend (clone and build)
setup-fdo-backend:
	@echo "Setting up FDO Go backend..."
	@if [ ! -d "../go-fdo" ]; then \
		echo "Cloning FDO Go repository..."; \
		git clone https://github.com/fido-device-onboard/go-fdo.git ../go-fdo; \
	fi
	@echo "Building FDO Go server..."
	cd ../go-fdo && go mod download && go build -o fdo-server ./cmd/server
	@echo "FDO Go backend setup complete"

# Full setup (backend + proxy)
setup: setup-fdo-backend build
	@echo "Creating necessary directories..."
	mkdir -p certs logs data
	@echo "Setup complete!"

# Run the proxy in basic mode
run:
	@echo "Starting FDO Server Proxy in basic mode..."
	./fdo-proxy -listen localhost:8080 -debug

# Run the proxy with passport integration (example)
run-with-passport:
	@echo "Starting FDO Server Proxy with passport integration..."
	./fdo-proxy \
		-listen localhost:8080 \
		-product-base-url https://cmulk1.cymanii.org:8443 \
		-commissioning-url http://cmulk1.cymanii.org:8000/create-commissioning-passport \
		-ca-cert ./certs/passport-service.pem \
		-client-cert ./certs/ucse-agent.crt \
		-client-key ./certs/ucse-agent.pem \
		-enable-product-passport \
		-owner-id test-owner \
		-debug

# Docker operations
docker-build:
	@echo "Building Docker image..."
	docker build -t fdo-server-proxy .

docker-run:
	@echo "Running FDO Go server in Docker..."
	@if [ ! -f "run-fdo-docker.sh" ]; then \
		echo "Creating Docker run script..."; \
		chmod +x scripts/setup.sh && ./scripts/setup.sh; \
	fi
	./run-fdo-docker.sh

docker-stop:
	@echo "Stopping Docker containers..."
	docker stop fdo-go-server 2>/dev/null || true
	docker rm fdo-go-server 2>/dev/null || true

# Docker Compose operations
compose-up:
	@echo "Starting full stack with Docker Compose..."
	docker-compose up -d

compose-down:
	@echo "Stopping Docker Compose stack..."
	docker-compose down

compose-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Run automated setup script
auto-setup:
	@echo "Running automated setup..."
	@if [ ! -f "scripts/setup.sh" ]; then \
		echo "Setup script not found. Please ensure scripts/setup.sh exists."; \
		exit 1; \
	fi
	chmod +x scripts/setup.sh && ./scripts/setup.sh

# Show help
help:
	@echo "FDO Server Proxy Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build              - Build the proxy executable"
	@echo "  test               - Run all tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  test-coverage-report - Run tests and generate HTML coverage report"
	@echo "  clean              - Remove build artifacts"
	@echo "  deps               - Install/update dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code (requires golangci-lint)"
	@echo "  setup-fdo-backend  - Clone and build FDO Go backend"
	@echo "  setup              - Full setup (backend + proxy + directories)"
	@echo "  auto-setup         - Run automated setup script"
	@echo "  run                - Run proxy in basic mode"
	@echo "  run-with-passport  - Run proxy with passport integration (example)"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run FDO Go server in Docker"
	@echo "  docker-stop        - Stop Docker containers"
	@echo "  compose-up         - Start full stack with Docker Compose"
	@echo "  compose-down       - Stop Docker Compose stack"
	@echo "  compose-logs       - Show Docker Compose logs"
	@echo "  help               - Show this help message" 