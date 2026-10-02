# Terraform Provider for Catena (ST2138)

A Terraform provider for managing Catena devices and services compatible with SMPTE ST2138.

## Features

- **Device Management**: Manage Catena/ST2138 device slots with `st2138_device`
- **gRPC and REST**: Use gRPC or the SMPTE REST API at `/st2138-api/v1`
- **Typed Parameters**: Configure scalars, arrays, structs, struct variants, and binary data payloads
- **File Payloads**: Supply binary payloads as base64 strings or read them from a file
- **Lifecycle Commands**: Optionally run commands on create and destroy, with status polling
- **Computed Snapshots**: Read writable parameters, full parameters, and commands back from the device

## Requirements

- Terraform >= 1.0
- Go >= 1.26.6 (for building from source)

## Installation

### Using OpenTofu Registry

```hcl
terraform {
  required_providers {
    st2138 = {
      source = "rossvideo/st2138"
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

The provider currently has no configuration arguments:

```hcl
provider "st2138" {}
```

The endpoint and transport are configured on each device's `network` block.

## Usage

### Creating a Device

```hcl
resource "st2138_device" "example" {
  name = "example-device"
  slot = 0

  network {
    address   = "localhost"
    port      = 6254
    transport = "grpc"
  }

  parameters = {
    counter = 1
  }
}
```

`network.transport` supports `grpc` and `rest`; REST uses the standardized SMPTE routes under `/st2138-api/v1`. Startup and shutdown command blocks are optional. See [docs/resources/device.md](docs/resources/device.md) for parameter types, binary payloads, command configuration, and lifecycle behavior.

## Development

### Project Structure

- `/internal/provider/` - Provider configuration
- `/internal/services/device/` - Device resource implementation
- `/internal/client/` - gRPC and REST client code
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

# Run tests serially and generate a raw Go coverage profile
go test ./... -coverprofile=coverage-all.out -covermode=atomic -count=1 -p=1 -parallel=1

# Exclude generated protobuf code from the reported project coverage
grep -v '/internal/genproto/' coverage-all.out > coverage.out

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Generate HTML coverage report (LCOV format)
genhtml lcov.info -o coverage_html
```

Open `coverage.html` or `coverage_html/index.html` in a browser to view a report.

**Current Coverage:**
- Root package: 100.0%
- `internal/client`: 81.2%
- `internal/client/params`: 100.0%
- `internal/datasources`: 93.5%
- `internal/provider`: 87.5%
- `internal/services/command`: 91.7%
- `internal/services/device`: 75.4%
- `internal/services/parameters`: 91.4%
- Handwritten project code statement coverage (generated protobuf excluded): 80.5%
- Handwritten project function coverage (generated protobuf excluded): 100.0%; no functions have zero coverage.

Coverage was measured with the serial test command above. `coverage-all.out` retains the unfiltered profile; `coverage.out` and `lcov.info` exclude generated protobuf code.

### Building

```bash
go build -o terraform-provider-st2138
```

### Releasing (OpenTofu Registry)

This repository includes a tag-driven GitHub Actions release workflow using GoReleaser.

1. Create and push a semantic version tag:

```bash
git tag v0.2.1
git push origin v0.2.1
```

2. GitHub Actions publishes release assets including:

- `terraform-provider-st2138_<version>_<os>_<arch>.zip`
- `terraform-provider-st2138_<version>_SHA256SUMS`

These are required for OpenTofu Registry version detection.

## Documentation

See the [`docs/`](docs/index.md) directory for provider and resource documentation.

See `/examples/` directory for usage examples.

## License

See LICENSE file for details.

## Contributing

Contributions welcome! Please follow existing code patterns and ensure tests pass before submitting PRs.
