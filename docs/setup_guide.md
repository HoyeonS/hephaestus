# Hephaestus Setup Guide

## Prerequisites

### Required Software
- Go 1.19 or later
- Protocol Buffers compiler (protoc)
- Make
- Git
- Docker (optional, for containerized deployment)

### System Requirements
- 2 CPU cores
- 4GB RAM
- 20GB disk space

## Installation

### 1. Install Go
```bash
# macOS (using Homebrew)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew install go

# Linux (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install golang-go

# Windows
# Download installer from https://golang.org/dl/
```

### 2. Install Protocol Buffers
```bash
# macOS
brew install protobuf

# Linux (Ubuntu/Debian)
sudo apt-get install protobuf-compiler

# Windows
# Download from https://github.com/protocolbuffers/protobuf/releases
```

### 3. Clone Repository
```bash
git clone https://github.com/HoyeonS/hephaestus.git
cd hephaestus
```

### 4. Install Dependencies
```bash
make deps
make proto-setup
```

### 5. Generate Protocol Buffer Code
```bash
make proto
```

### 6. Build Project
```bash
make build
```

## Configuration

### 1. Create Configuration File
Create a file named `config.yaml` in the project root:

```yaml
# Basic Configuration
log_level: info
log_output: stdout
node_id: your-node-id
mode: suggest  # or deploy

# Remote Repository Configuration
repository:
  provider: your-provider  # e.g., gitlab, bitbucket
  owner: your-repo-owner
  name: your-repo-name
  token: your-access-token
  base_path: /path/to/repo
  branch: main

# Model Service Configuration
model:
  provider: your-provider
  api_key: your-model-api-key
  base_url: https://api.example.com
  timeout: 30

# Logging Configuration
logging:
  level: info
  format: json
  output: stdout

# Metrics Configuration
metrics:
  enabled: true
  port: 9090
```

### 2. Environment Variables
You can also configure the service using environment variables:

```bash
# Basic Configuration
export LOG_LEVEL=info
export LOG_OUTPUT=stdout
export NODE_ID=your-node-id
export MODE=suggest

# Remote Repository Configuration
export REPOSITORY_PROVIDER=your-provider
export REPOSITORY_TOKEN=your-access-token
export REPOSITORY_OWNER=your-repo-owner
export REPOSITORY_NAME=your-repo-name
export REPOSITORY_BRANCH=main

# Model Service Configuration
export MODEL_PROVIDER=your-provider
export MODEL_API_KEY=your-model-api-key
```

## Running the Service

### 1. Start the Server
```bash
./hephaestus server
```

### 2. Run with Docker
```bash
# Build Docker image
docker build -t hephaestus .

# Run container
docker run -d \
  --name hephaestus \
  -p 50051:50051 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  hephaestus
```

## Verification

### 1. Check Service Status
```bash
curl http://localhost:50051/health
```

### 2. Run Tests
```bash
make test          # Run all tests
make unit-test     # Run unit tests only
make integration-test  # Run integration tests
```

## Troubleshooting

### Common Issues

1. **Connection Issues**
   - Check network connectivity
   - Verify port availability
   - Check firewall settings

2. **Authentication Errors**
   - Verify access tokens
   - Check permissions
   - Validate configuration

3. **Resource Issues**
   - Monitor system resources
   - Check disk space
   - Verify memory usage

### Logging

1. **View Logs**
   ```bash
   # View service logs
   tail -f /var/log/hephaestus.log

   # View Docker logs
   docker logs -f hephaestus
   ```

2. **Change Log Level**
   ```bash
   # Update config.yaml
   log_level: debug

   # Or use environment variable
   export LOG_LEVEL=debug
   ```

## Maintenance

### 1. Backup Configuration
```bash
cp config.yaml config.yaml.backup
```

### 2. Update Service
```bash
git pull
make build
systemctl restart hephaestus
```

### 3. Monitor Resources
```bash
# Check system resources
top -p $(pgrep hephaestus)

# Check disk usage
df -h /var/log/hephaestus
```

## Support

For additional support:
1. Check the documentation
2. Review issue tracker
3. Contact the maintainers 