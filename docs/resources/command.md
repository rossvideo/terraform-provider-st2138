# st2138_command Resource

Manages a single Catena/ST2138 command execution.

This resource defines a reusable command payload. `st2138_device` can consume these through `startup_commands` and `shutdown_commands` blocks.

## Example

```hcl
resource "st2138_command" "start_ooe_command" {
    command         = "/fib_start"
    timeout_seconds  = 5
}

resource "st2138_command" "set_ooe_command_with_value" {
    command         = "/fib_set"
    value           = 0
    timeout_seconds = 5
}

resource "st2138_command" "set_ooe_command_check_for_success" {
    command                   = "/fib_set"
    value                     = 1
    timeout_seconds           = 5
    status_foid               = "number_example"
    status_success_comparator = "eq"
    status_success_value      = 1
}
```

## Argument Reference

### Required

- `command` (String): OID of the command to execute.

### Optional

- `value` (Dynamic): Value passed with the command when the OID expects one.
- `status_foid` (String): Parameter OID to poll after command execution.
- `status_success_value` (String): Expected value used by the comparator.
- `status_success_comparator` (String): One of `eq`, `ne`, `gt`, `lt`, `ge`, `le`.
- `timeout_seconds` (Number): Poll timeout in seconds. Default is 5.

## Behavior

1. Stores a reusable command configuration in Terraform state.
2. Sets `timeout_seconds` to 5 when omitted.
3. Allows `st2138_device` lifecycle blocks to reference command resources.

## Notes

- This resource does not execute commands by itself.
- Command execution happens when referenced from `st2138_device.startup_commands` or `st2138_device.shutdown_commands`.
