# Remote gRPC Device Example

This example shows how to configure a remote device accessed via gRPC.

## Usage

Update the `address` and `port` variables to match your device:

```bash
terraform init
terraform plan -var="device_address=192.168.1.100"
terraform apply -var="device_address=192.168.1.100"
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

# Variables for device connection
variable "device_address" {
  description = "Remote device IP address or hostname"
  type        = string
  default     = "localhost"
}

variable "device_port" {
  description = "Remote device gRPC port"
  type        = number
  default     = 6254
}

variable "use_tls" {
  description = "Use TLS for gRPC connection"
  type        = bool
  default     = false
}

# Remote device configuration
resource "st2138_device" "remote" {
  name        = "remote-device"
  device_type = "remote-grpc"
  address     = var.use_tls ? "grpcs://${var.device_address}" : var.device_address
  port        = var.device_port
  slot        = 0

  # Configure device parameters
  params_map = {
    "/network/hostname"   = var.device_address
    "/network/port"       = tostring(var.device_port)
    "/system/mode"        = "production"
    "/audio/input/0/name" = "Main Audio"
    "/video/input/0/name" = "Main Video"
  }

  # Monitor device status
  device_status {
    oid         = "/system/state"
    ready_value = "online"
  }
}

# Outputs
output "remote_device_id" {
  value = st2138_device.remote.id
}

output "connection_endpoint" {
  value = "${st2138_device.remote.address}:${st2138_device.remote.port}"
}

output "current_status" {
  value       = st2138_device.remote.status_value
  description = "Current device status from /system/state"
}
```

## What This Does

1. Connects to a remote device via gRPC
2. Optionally enables TLS with `grpcs://` scheme
3. Configures network and I/O parameters
4. Monitors device status at `/system/state`
5. Outputs connection details and current status

## TLS Example

To connect with TLS:

```bash
terraform apply -var="use_tls=true" -var="device_address=secure.example.com"
```

## Notes

- Ensure the remote device is accessible on the specified port
- Default gRPC port is 6254
- Firewall must allow gRPC traffic
- TLS requires proper certificate configuration on the device
