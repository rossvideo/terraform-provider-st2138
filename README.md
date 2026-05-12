# Terraform Provider for Catena (ST2138)

A Terraform provider for managing Catena devices and services compatible with SMPTE ST2138.

## Features

- **Device Management**: Create, read, update, and delete Catena devices
- **Dual Transport Support**: Manage devices via both gRPC and REST/HTTP endpoints
- **Remote gRPC Support**: Manage remote devices via gRPC endpoints
- **REST/HTTP Support**: Access devices through REST API endpoints with streaming
- **Parameter Configuration**: Set device parameters via OID-based configuration
- **Device Status Monitoring**: Poll device status and wait for readiness conditions
- **Container Support**: Docker-based device deployment (experimental)

## Requirements

- Terraform >= 1.0
- Go >= 1.25 (for building from source)

## Installation

### Using OpenTofu Registry (Coming Soon)

```hcl
terraform {
  required_providers {
    catena = {
      source = "registry.opentofu.org/rossvideo/st2138"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/rossvideo/terraform-provider-st2138.git
cd terraform-provider-st2138
go build -o terraform-provider-st2138
```

## Configuration

Provider configuration example:

```hcl
# gRPC transport (default)
provider "catena" {
  endpoint       = "localhost:6254"  # gRPC service endpoint
  transport      = "grpc"            # Transport protocol: grpc, http, https, rest
  devices_dir    = "../devices"      # Directory containing device definitions
  executables_dir = "../devices"     # Alias for devices_dir
}

# REST/HTTP transport
provider "catena" {
  endpoint       = "http://device.example.com:443"  # REST API endpoint
  transport      = "http"                            # Transport protocol: http, https, rest
  devices_dir    = "../devices"
}
```

### Provider Arguments

- `endpoint` (Optional) - Service endpoint in `host:port` or `scheme://host:port` format
- `transport` (Optional) - Transport protocol:
  - `grpc` (default) - Binary gRPC protocol
  - `http` - REST API over HTTP
  - `https` - REST API over HTTPS (TLS-secured)
  - `rest` - Auto-detect HTTP/HTTPS from endpoint scheme
- `devices_dir` (Optional) - Base directory for device-type files
- `executables_dir` (Optional) - Alias for `devices_dir`

## Usage

### Creating a Device

```hcl
resource "catena_device" "example" {
  slot        = 1
  name        = "my-device"
  device_type = "remote-grpc"
  address     = "192.168.1.100"
  port        = 6254

  params_map = {
    "/input/0/name"  = "Input 1"
    "/input/1/name"  = "Input 2"
  }

  device_status {
    oid         = "/status/ready"
    ready_value = "true"
  }
}
```

### Parameters

- `slot` (Required) - Device slot ID for gRPC calls
- `name` (Optional) - Human-readable device name
- `device_type` (Optional) - Device type identifier (e.g., `pat2mxl`, `remote-grpc`)
- `address` (Optional) - Remote device address (required for `remote-grpc`)
- `port` (Optional) - Remote device port (required for `remote-grpc`)
- `params_map` (Optional) - Map of OID to value pairs
- `params` (Optional) - List of OID/value pair blocks for structured configuration
- `start_command` (Optional) - Command to run after device startup
- `stop_command` (Optional) - Command to run before device deletion
- `device_status` (Optional) - Status polling configuration block
- `apply_all` (Optional) - When true, always apply all params on every update

### Computed Values

- `id` - Unique device identifier
- `container_id` - Docker container ID (if applicable)
- `status_value` - Current polled status value

## Development

### Project Structure

- `/internal/provider/` - Provider configuration
- `/internal/services/device/` - Device resource implementation
- `/internal/client/` - gRPC client code
- `/internal/genproto/` - Generated protobuf files
- `/examples/` - Terraform configuration examples
- `/docs/` - API documentation

### Running Tests

```bash
go test ./...
```

### Test Coverage

Generate coverage reports in multiple formats:

```bash
# Run tests and generate lcov.info
./test.sh

# View coverage summary
go test ./... -cover

# Generate HTML coverage report (Go native)
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Generate HTML coverage report (LCOV format)
genhtml lcov.info -o coverage_html
open coverage_html/index.html
```

**Current Coverage:**
- Client package: 72.3%
- Params package: 86.2%
- Overall project: 45.7%

The `test.sh` script generates both `coverage.out` (Go format) and `lcov.info` (LCOV format) for compatibility with various coverage visualization tools.

### Building

```bash
go build -o terraform-provider-st2138
```

### Releasing (OpenTofu Registry)

This repository includes a tag-driven GitHub Actions release workflow using GoReleaser.

1. Create and push a semantic version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

2. GitHub Actions publishes release assets including:

- `terraform-provider-st2138_<version>_<os>_<arch>.zip`
- `terraform-provider-st2138_<version>_SHA256SUMS`

These are required for OpenTofu Registry version detection.

## Documentation

See `/docs/` directory for comprehensive provider and resource documentation.

See `/examples/` directory for usage examples.

## License

See LICENSE file for details.

## Contributing

Contributions welcome! Please follow existing code patterns and ensure tests pass before submitting PRs.
