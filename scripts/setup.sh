#!/bin/bash

# FDO Server Proxy Setup Script
# This script automatically sets up the FDO Go backend and builds the proxy

set -e  # Exit on any error

echo "🚀 Setting up FDO Server Proxy..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please install Go 1.21 or later."
        exit 1
    fi
    
    print_success "Go version $GO_VERSION is compatible"
    
    # Check if Git is installed
    if ! command -v git &> /dev/null; then
        print_error "Git is not installed. Please install Git."
        exit 1
    fi
    
    print_success "Git is available"
}

# Setup FDO Go backend
setup_fdo_backend() {
    print_status "Setting up FDO Go backend..."
    
    FDO_GO_PATH="../go-fdo"
    
    if [ -d "$FDO_GO_PATH" ]; then
        print_warning "FDO Go repository already exists at $FDO_GO_PATH"
        read -p "Do you want to update it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Updating FDO Go repository..."
            cd "$FDO_GO_PATH"
            git pull origin main
            cd - > /dev/null
        fi
    else
        print_status "Cloning FDO Go repository..."
        git clone https://github.com/fido-device-onboard/go-fdo.git "$FDO_GO_PATH"
    fi
    
    # Build FDO Go server
    print_status "Building FDO Go server..."
    cd "$FDO_GO_PATH"
    go mod download
    go build -o fdo-server ./cmd/server
    cd - > /dev/null
    
    print_success "FDO Go backend setup complete"
}

# Build the proxy
build_proxy() {
    print_status "Building FDO Server Proxy..."
    
    # Install dependencies
    go mod tidy
    
    # Build the proxy
    go build -o fdo-proxy ./cmd/server
    
    print_success "FDO Server Proxy built successfully"
}

# Create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p certs
    mkdir -p logs
    mkdir -p data
    
    print_success "Directories created"
}

# Create example configuration
create_example_config() {
    print_status "Creating example configuration..."
    
    cat > example-config.sh << 'EOF'
#!/bin/bash
# Example configuration for FDO Server Proxy

# Basic usage
./fdo-proxy -listen localhost:8080 -debug

# With passport service integration
./fdo-proxy \
  -listen localhost:8080 \
  -product-base-url https://cmulk1.cymanii.org:8443 \
  -commissioning-url http://cmulk1.cymanii.org:8000/create-commissioning-passport \
  -ca-cert ./certs/passport-service.pem \
  -client-cert ./certs/ucse-agent.crt \
  -client-key ./certs/ucse-agent.pem \
  -enable-product-passport \
  -owner-id your-owner-id \
  -debug
EOF

    chmod +x example-config.sh
    print_success "Example configuration created: example-config.sh"
}

# Create Docker run script
create_docker_script() {
    print_status "Creating Docker run script..."
    
    cat > run-fdo-docker.sh << 'EOF'
#!/bin/bash
# Script to run FDO Go server in Docker

# Stop existing container if running
docker stop fdo-go-server 2>/dev/null || true
docker rm fdo-go-server 2>/dev/null || true

# Pull latest image
docker pull fidoalliance/go-fdo:latest

# Run FDO Go server
docker run -d \
  --name fdo-go-server \
  -p 8081:8081 \
  -v $(pwd)/data:/app/data \
  fidoalliance/go-fdo:latest \
  -http localhost:8081 \
  -db /app/data/fdo-backend.db

echo "FDO Go server started in Docker container"
echo "Access it at: http://localhost:8081"
EOF

    chmod +x run-fdo-docker.sh
    print_success "Docker run script created: run-fdo-docker.sh"
}

# Main setup function
main() {
    echo "=========================================="
    echo "FDO Server Proxy Setup"
    echo "=========================================="
    
    check_prerequisites
    setup_fdo_backend
    build_proxy
    create_directories
    create_example_config
    create_docker_script
    
    echo ""
    echo "=========================================="
    print_success "Setup complete!"
    echo "=========================================="
    echo ""
    echo "Next steps:"
    echo "1. Place your certificates in the 'certs' directory"
    echo "2. Run the proxy: ./fdo-proxy -listen localhost:8080 -debug"
    echo "3. Or use Docker: ./run-fdo-docker.sh"
    echo "4. See example-config.sh for more options"
    echo ""
    echo "For more information, see the README.md file"
}

# Run main function
main "$@" 