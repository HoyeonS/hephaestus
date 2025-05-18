# Hephaestus Setup Guide

This guide will help you set up Hephaestus for both development and production use.

## Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (protoc) v3.15.0 or later
- Docker (optional, for containerized deployment)
- Git
- Make (for build automation)

## Installation Steps

### 1. Install Dependencies

#### For macOS:
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Install Protocol Buffers
brew install protobuf

# Install gRPC tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### For Linux (Ubuntu/Debian):
```bash
# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Protocol Buffers
sudo apt-get update
sudo apt-get install -y protobuf-compiler

# Install gRPC tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 2. Clone the Repository

```bash
git clone https://github.com/HoyeonS/hephaestus.git
cd hephaestus
```

### 3. Set Up Development Environment

```bash
# Add GOPATH to your environment if not already done
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Install project dependencies
go mod download

# Generate Protocol Buffer code
make proto
```

### 4. Build the Project

```bash
# Build all components
make build

# Run tests
make test

# Build Docker image (optional)
make docker-build
```

## Configuration

### 1. Environment Variables

Create a `.env` file in the project root:

```env
# Server Configuration
HEPHAESTUS_PORT=50051
HEPHAESTUS_HOST=localhost
HEPHAESTUS_MODE=development

# GitHub Configuration (if using GitHub integration)
GITHUB_TOKEN=your_github_token
GITHUB_OWNER=your_github_username
GITHUB_REPO=your_repository_name

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Security Configuration
TLS_ENABLED=false
TLS_CERT_PATH=/path/to/cert
TLS_KEY_PATH=/path/to/key
```

### 2. Application Configuration

Create a `config.yaml` file:

```yaml
server:
  host: localhost
  port: 50051
  mode: development

github:
  enabled: true
  owner: your_github_username
  repo: your_repository_name

logging:
  level: info
  format: json
  output: stdout

security:
  tls:
    enabled: false
    certPath: /path/to/cert
    keyPath: /path/to/key
```

## Running the Application

### Development Mode

```bash
# Run the server
make run

# Run with specific configuration
HEPHAESTUS_CONFIG_PATH=./config.yaml make run
```

### Production Mode

```bash
# Using Docker
docker run -p 50051:50051 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -e HEPHAESTUS_MODE=production \
  hephaestus:latest

# Using systemd
sudo cp deployment/hephaestus.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable hephaestus
sudo systemctl start hephaestus
```

## Verification

### 1. Check Server Status

```bash
# Check if server is running
curl http://localhost:50051/health

# Check server version
curl http://localhost:50051/version
```

### 2. Run Example Client

```bash
# Run the example initialization client
go run examples/initialization/main.go
```

## Troubleshooting

### Common Issues

1. **Protocol Buffer Generation Fails**
   - Ensure protoc is installed correctly
   - Check GOPATH is set properly
   - Verify protoc-gen-go and protoc-gen-go-grpc are installed

2. **Server Won't Start**
   - Check port availability
   - Verify configuration file paths
   - Check log files for errors

3. **GitHub Integration Issues**
   - Verify GitHub token permissions
   - Check network connectivity
   - Validate repository access

### Logging

- Logs are written to stdout by default
- Check system logs: `journalctl -u hephaestus`
- Enable debug logging by setting `LOG_LEVEL=debug`

## Development Tools

### Recommended VSCode Extensions

- Go extension
- Protocol Buffers extension
- Docker extension
- YAML extension

### Code Generation

```bash
# Generate mocks
make generate-mocks

# Generate Protocol Buffer code
make proto

# Generate API documentation
make docs
```

### Testing

```bash
# Run all tests
make test

# Run specific tests
go test ./pkg/hephaestus/...

# Run with coverage
make test-coverage
```

## Next Steps

1. Read the [API Documentation](./api.md)
2. Explore [Example Code](../examples/)
3. Review [Architecture Documentation](./HLD.md)
4. Join the [Community](../CONTRIBUTING.md)

## Support

- Create an issue on GitHub
- Join our Discord community
- Check the FAQ
- Contact the maintainers

## Security Considerations

1. Always use HTTPS/TLS in production
2. Secure your GitHub tokens
3. Follow the security guidelines in the documentation
4. Regularly update dependencies

## Maintenance

1. Regular Updates
   ```bash
   # Update dependencies
   go get -u ./...
   
   # Update Protocol Buffers
   make proto
   ```

2. Monitoring
   - Set up Prometheus metrics
   - Configure alerting
   - Monitor system resources

3. Backup
   - Regularly backup configuration
   - Document deployment settings
   - Version control all changes 