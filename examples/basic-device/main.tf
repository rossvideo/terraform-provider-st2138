# Basic Device Example
terraform {
  required_providers {
    st2138 = {
      source = "rossvideo/st2138"
      version = "0.1.0"
    }
  }
}


provider "st2138" {}

resource "st2138_parameters" "basic_params" {  # Set a few basic parameters
  parameters = {
    "name"        = "Basic Example Device"
    "description" = "Minimal configuration example"
    "enabled"     = "true"
  }
}
resource "st2138_parameters" "basic_params_file" {  # Set a few basic parameters
  parameters_file = "parameters.stpm"
}

resource "st2138_device" "basic" {
  name = "basic-device"
  slot = 0
  network {
    address   = "localhost"
    port      = 6254
    transport = "grpc"
    tls       = false
  }
  override_param_values_on_update = false
  parameters = st2138_parameters.basic_params.parameters
}

# Output the device ID and status
output "device_id" {
  value = st2138_device.basic.id
}

output "device_status" {
  value = st2138_device.basic.status_value
}
