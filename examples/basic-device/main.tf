# Basic Device Example

This example demonstrates the simplest device configuration.

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Configuration

```hcl
terraform {
  required_providers {
    st2138 = {
      source = "rossvideo/st2138"
    }
  }
}

provider "st2138" {}

resource "st2138_device" "basic" {
  name = "basic-device"
  slot = 0

  # Set a few basic parameters
  params_map = {
    "/system/name"        = "Basic Example Device"
    "/system/description" = "Minimal configuration example"
    "/system/enabled"     = "true"
  }
}

# Output the device ID and status
output "device_id" {
  value = st2138_device.basic.id
}

output "device_status" {
  value = st2138_device.basic.status_value
}
```

## What This Does

1. Creates a device named "basic-device" on slot 0
2. Sets three system parameters
3. Outputs the device ID and current status

## Next Steps

- See `remote-device/` for remote gRPC device example
- See `device-lifecycle/` for start/stop command example
- See `multi-parameter/` for advanced parameter configuration
