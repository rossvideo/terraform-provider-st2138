
# This example will
# on create: will add the device to the tofu inventory, and set all parameters to the values specified, then runs the startup_command block
# on update: sets only the parameters whose values changed
# on delete: runs the shutdown_command block, then deletes the device from the tofu inventory


terraform {
  required_providers {
    st2138 = {
      source = "rossvideo/st2138"
      version = "0.1.0"
    }
  }
}

provider "st2138" {}

resource "st2138_command" "start_counter_command" {
    command                 = "start"
    status_foid              = "running"
    status_success_comparator = "ne"
    status_success_value     = "0"
    timeout_seconds          = 5
}
resource "st2138_command" "stop_counter_command" {
    command                 = "stop"
    status_foid              = "running"
    status_success_comparator = "ne"
    status_success_value     = "1"
    timeout_seconds          = 5
}

resource "st2138_command" "reset_counter_command" {
    command                 = "reset"
    status_foid              = "counter"
    status_success_comparator = "eq"
    status_success_value     = "0"
    timeout_seconds          = 5
}

resource "st2138_device" "one_of_everything_slot0" {
  depends_on = [st2138_device.one_of_everything_slot1,st2138_device.one_of_everything_slot2]
  name                            = "One of Everything"
  slot                            = 0
  network {
    address                         = "localhost"
    port                            = 6254
    transport                       = "grpc"
    tls                             = false
  }

  
  parameters = [
    {
      counter = 1

    },
  ]

  startup_commands {
    commands = [st2138_command.reset_counter_command, st2138_command.start_counter_command]
  }

  shutdown_commands {
    commands = [st2138_command.stop_counter_command, st2138_command.reset_counter_command]
  }
  
}

resource "st2138_device" "one_of_everything_slot1" {
  name                            = "One of Everything"
  slot                            = 1
  network {
    address                         = "localhost"
    port                            = 6254
    transport                       = "grpc"
    tls                             = false
  }

  
  parameters = [
    {
      brightness = 100
      contrast = 51
      saturation = 55
      resolution = "1920x1400"
    },
  ]
}

resource "st2138_device" "one_of_everything_slot2" {
  name                            = "One of Everything"
  slot                            = 2
  network {
    address                         = "localhost"
    port                            = 6254
    transport                       = "grpc"
    tls                             = false
  }

  
  parameters = [
    {
     sample_string_array = ["alpha","bravo","charlie"]
     sample_float_array = [1.0,1.1,2.0]
     sample_int_array = [5, 4, 3, 2]
     device_name ="tofu controled demo device"
     muted = 1
     volume = 100
     sample_struct_variant_array = [
      {
        nested_struct = {
          struct_variant_type = "int_kind"
          value = 10
        }
      },
      {
        nested_struct = {
          struct_variant_type = "string_kind"
          value = "hello"
        }
      }
      
     ]
     sample_struct_array = [
      {
        nested_struct = {
          label = "entry_a_1"
          count = 10
        }
      },
      {
        nested_struct = {
          label = "entry_b_1"
          count = 11
        }
      },
     ]
     sample_float = 2.821
     struct_example = {
        nested_struct = {
          number = 1
          text = "Slot 2 struct that was set by tofu"
        }
     }
     sample_struct_variant = {
       nested_struct = {
         struct_variant_type = "int_kind"
         value = 10
       }
     }
     
     sample_binary = {
      data_payload ={
        "metadata": {},
        "digest": "",
        "payload_encoding": "UNCOMPRESSED",
        "payload": "yv66vg=="
     }}

    #  sample_binary = {
    #   data_payload = {
    #     metadata         = {}
    #     payload_encoding = "UNCOMPRESSED"
    #     payload_file     = "${path.module}/payload.bin"
    #   }
    # }
    },
  ]
}


# Output writable parameters with native OpenTofu values where possible.
# JSON-looking strings from the provider are decoded into numbers, lists, and maps.
output "device_params" {
  description = "writable parameters for the configured slot with decoded values where possible"
  value = {
    for foid, raw in st2138_device.one_of_everything_slot0.parameters_out :
    foid => try(jsondecode(raw), raw)
  }
}
# same as above but shows all params, including read only ones. Useful for debugging.
output "device_full_params" {
  description = "all parameters for the configured slot with decoded values where possible"
  value = {
    for foid, raw in st2138_device.one_of_everything_slot0.full_parameters_out :
    foid => try(jsondecode(raw), raw)
  }
}
# Output the commands with native OpenTofu values where possible.
output "device_commands" {
  description = "commands for the configured slot with decoded values where possible"
  value = {
    for foid, raw in st2138_device.one_of_everything_slot0.commands_out :
    foid => try(jsondecode(raw), raw)
  }

}