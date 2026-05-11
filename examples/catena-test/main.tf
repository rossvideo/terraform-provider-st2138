
# This example will
# on create: will add the device to the tofu inventory, and set all parameters to the values specified, then runs the startup_command block
# on update: do nothing, since override_param_values_on_update is false 
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

resource "st2138_command" "start_ooe_command" {
    command                 = "/fib_start"
    status_foid              = "number_example"
    status_success_comparator = "ne"
    status_success_value     = "0"
    timeout_seconds          = 5
}
resource "st2138_command" "stop_ooe_command" {
    command                 = "/fib_stop"
    timeout_seconds          = 5
}

resource "st2138_command" "set_0_ooe_command" {
    command                 = "/fib_set"
    value                   = { int32_value = 0 }
    status_foid              = "number_example"
    status_success_comparator = "eq"
    status_success_value     = "0"
    timeout_seconds          = 5
}

resource "st2138_device" "one_of_everything" {
  name                            = "One of Everything"
  slot                            = 0
  network {
    address                         = "localhost"
    port                            = 6254
    transport                       = "grpc"
    tls                             = false
  }

  override_param_values_on_update = false
  parameters = [
    {
      "authz_admin"     = "You have st2138:adm scope!"
      "authz_configure" = "You have st2138:cfg scope!"
      "authz_monitor"   = "You have st2138:mon scope!"
      "authz_operate"   = "You have st2138:op scope!"
      "constraint_examples" = {
        "float32_range"              = 10
        "float_array_range"          = [11.5, 22.5, 33.5, 44.5, ]
        "int32_choice"               = 1
        "int32_range"                = 6
        "int_array_choice"           = [0, 1, 0, 0, 1, ]
        "int_array_range"            = [0, 2, 4, 6, 8, 10, ]
        "string_array_choice"        = ["a", "b", "a", ]
        "string_array_length"        = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", ]
        "string_choice"              = "a"
        "string_length"              = "Hello worl"
        "string_string_array_choice" = ["<#FF0000>", "<#00FF00>", "<#0000FF>", ]
        "string_string_choice"       = "<#FF0000>"
      }
      "counter"        = 1
      "float_array"    = [1.1, 2.2, 3.3, 4.4, ]
      "float_example"  = 0
      "menu_button"    = 0
      "number_array"   = [1, 2, 3, 4, ]
      "number_example" = 0
      "ref_struct"     = ""
      "string_array"   = ["one", "two", "three", "four", "five", ]
      "string_example" = "Hello World"
      "struct_array" = [
        {
          "nested_struct" = {
            "num_1" = 1
            "num_2" = 2
          }
        },
        {
          "nested_struct" = {
            "num_1" = 3
            "num_2" = 4
          }
        },
      ]
      "struct_example" = {
        "nested_struct" = {
          "num_1" = 1
          "num_2" = 2
        }
      }
    },
  ]

  startup_commands {
    commands = [st2138_command.set_0_ooe_command, st2138_command.start_ooe_command]
  }

  shutdown_commands {
    commands = [st2138_command.stop_ooe_command, st2138_command.set_0_ooe_command]
  }
  
}

# Output writable parameters with native OpenTofu values where possible.
# JSON-looking strings from the provider are decoded into numbers, lists, and maps.
output "device_params" {
  description = "writable parameters for the configured slot with decoded values where possible"
  value = {
    for foid, raw in st2138_device.one_of_everything.parameters_out :
    foid => try(jsondecode(raw), raw)
  }
}
# same as above but shows all params, including read only ones. Useful for debugging.
output "device_full_params" {
  description = "all parameters for the configured slot with decoded values where possible"
  value = {
    for foid, raw in st2138_device.one_of_everything.full_parameters_out :
    foid => try(jsondecode(raw), raw)
  }
}
# Output the commands with native OpenTofu values where possible.
output "device_commands" {
  description = "commands for the configured slot with decoded values where possible"
  value = {
    for foid, raw in st2138_device.one_of_everything.commands_out :
    foid => try(jsondecode(raw), raw)
  }

}