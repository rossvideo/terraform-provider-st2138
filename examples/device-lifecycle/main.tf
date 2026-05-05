# Device Lifecycle Management Example

This example demonstrates start/stop command handling and device status monitoring.

## Usage

```bash
terraform init
terraform apply
# Device starts and waits for ready state
terraform destroy
# Device stops and waits for shutdown
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

resource "st2138_device" "managed" {
  name = "managed-device"
  slot = 0

  # Lifecycle commands
  start_command = "/app/scripts/startup.sh"
  stop_command  = "/app/scripts/shutdown.sh"

  # Status monitoring
  device_status {
    oid         = "/system/operational_state"
    ready_value = "running"
  }

  # Initial configuration applied after device starts
  params_map = {
    "/system/name"             = "Managed Device Example"
    "/system/auto_start"       = "false"  # We control startup via start_command
    "/monitoring/health_check" = "true"
    "/logs/level"              = "info"
  }

  # Additional parameter blocks for specific settings
  params {
    oid   = "/config/environment"
    value = "production"
  }

  params {
    oid   = "/config/debug_mode"
    value = "false"
  }
}

# Example with multiple devices managed together
resource "st2138_device" "primary" {
  name          = "primary-device"
  slot          = 0
  start_command = "/scripts/start-primary.sh"
  stop_command  = "/scripts/stop-primary.sh"

  device_status {
    oid         = "/status/state"
    ready_value = "active"
  }

  params_map = {
    "/role" = "primary"
  }
}

resource "st2138_device" "secondary" {
  name          = "secondary-device"
  slot          = 1
  start_command = "/scripts/start-secondary.sh"
  stop_command  = "/scripts/stop-secondary.sh"

  device_status {
    oid         = "/status/state"
    ready_value = "active"
  }

  params_map = {
    "/role"           = "secondary"
    "/primary_device" = st2138_device.primary.id
  }

  # Secondary depends on primary being ready
  depends_on = [st2138_device.primary]
}

# Outputs
output "primary_status" {
  value = st2138_device.primary.status_value
}

output "secondary_status" {
  value = st2138_device.secondary.status_value
}
```

## Lifecycle Flow

### On Apply (Create)

1. Terraform creates the device resource
2. Sets initial parameters from `params_map` and `params`
3. Waits 3 seconds
4. Executes `start_command` via gRPC
5. Polls `device_status.oid` every second
6. Waits up to 60 seconds for value to equal `ready_value`
7. Configuration complete when device reports ready

### On Destroy (Delete)

1. Terraform initiates device deletion
2. Executes `stop_command` via gRPC
3. Waits 5 seconds
4. Polls `device_status.oid` every second
5. Waits up to 60 seconds for value to differ from `ready_value`
6. Deletion complete when device leaves ready state

## Notes

- Start command runs AFTER parameters are set
- Stop command runs BEFORE device is deleted
- Status polling has 60-second timeout
- Device must respond to gRPC commands
- Scripts should return 0 on success
