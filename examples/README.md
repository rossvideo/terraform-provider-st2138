# Terraform Provider ST2138 Examples

This directory contains example configurations for the ST2138 Terraform provider.

## Available Examples

### [basic-device/](./basic-device/)
Minimal device configuration demonstrating:
- Simple device creation
- Basic parameter setting with `params_map`
- Output usage

**Use this when**: Getting started with the provider.

### [remote-device/](./remote-device/)
Remote gRPC device configuration demonstrating:
- Connecting to remote devices
- TLS/non-TLS connections
- Variable usage for flexible configuration
- Status monitoring

**Use this when**: Managing devices on remote hosts.

### [device-lifecycle/](./device-lifecycle/)
Lifecycle management demonstrating:
- Start and stop command execution
- Device status monitoring
- Coordinated startup of multiple devices
- Dependency management between devices

**Use this when**: Devices require startup/shutdown procedures.

### [multi-parameter/](./multi-parameter/)
Advanced parameter management demonstrating:
- Bulk parameter configuration
- Computed parameter values using `locals`
- Conditional configuration
- Parameter precedence
- The `apply_all` flag

**Use this when**: Managing complex device configurations.

## Quick Start

1. Choose an example directory
2. Navigate to it: `cd basic-device/`
3. Initialize Terraform: `terraform init`
4. Review the plan: `terraform plan`
5. Apply the configuration: `terraform apply`

## Provider Configuration

All examples use the ST2138 provider. The minimal provider configuration is:

```hcl
terraform {
  required_providers {
    st2138 = {
      source = "rossvideo/st2138"
    }
  }
}

provider "st2138" {}
```

## Common Patterns

### Setting Parameters

Use `params_map` for bulk configuration:
```hcl
params_map = {
  "/system/name" = "My Device"
  "/config/mode" = "production"
  "/inputs/0/enabled" = "true"
}
```

Use `params` blocks for individual parameters or to override `params_map`:
```hcl
params {
  oid   = "/advanced/setting"
  value = "custom-value"
}
```

### Remote Device Connection

```hcl
resource "st2138_device" "remote" {
  device_type = "remote-grpc"
  address     = "192.168.1.100"  # or "grpcs://secure.host.com" for TLS
  port        = 6254
  slot        = 0
  # ...
}
```

### Device Lifecycle

```hcl
resource "st2138_device" "managed" {
  start_command = "/path/to/startup.sh"
  stop_command  = "/path/to/shutdown.sh"
  
  device_status {
    oid         = "/system/state"
    ready_value = "running"
  }
  # ...
}
```

## Parameter Value Types

Values in `params_map` and `params` are automatically typed:
- Numbers: `"123"`, `"45.67"`
- Booleans: `"true"`, `"false"`
- Strings: `"any text"`

## Testing Examples

Most examples can be tested without actual hardware by using:
1. Mock gRPC servers
2. Local development devices
3. The provider's built-in validation

Modify connection details in each example to match your environment.

## Documentation

See [../docs/](../docs/) for complete provider and resource documentation.

## Support

- GitHub Issues: https://github.com/rossvideo/terraform-provider-st2138/issues
- Provider Documentation: https://registry.terraform.io/providers/rossvideo/st2138/latest/docs
