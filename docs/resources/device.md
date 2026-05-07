# st2138_device Resource

Manages a Catena/ST2138 device.


## Example

```hcl
resource "st2138_device" "one_of_everything" {
  name = "One of Everything"
  slot = 0

  network {
    address   = "localhost"
    port      = 6254
    transport = "grpc"
    tls       = false
  }

  override_param_values_on_update = false

  parameters = [
    {
      counter        = 1
      number_example = 0
      string_example = "Hello World"
      float_array    = [1.1, 2.2, 3.3]
      struct_example = {
        nested_struct = {
          num_1 = 1
          num_2 = 2
        }
      }
    }
  ]

  startup_command {
    commands                  = ["/fib_start"]
    values                    = []
    status_foid               = "number_example"
    status_success_comparator = "ne"
    status_success_value      = "0"
    timeout_seconds           = 5
  }

  shutdown_command {
    commands                  = ["/fib_stop", "/fib_set"]
    values                    = [null, { int32_value = 0 }]
    status_foid               = "number_example"
    status_success_comparator = "eq"
    status_success_value      = "0"
    timeout_seconds           = 5
  }
}
```

See full working example in [examples/catena-test/main.tf](examples/catena-test/main.tf).

## Argument Reference

### Required

- `slot` (Number): Device slot managed by this resource.
- `network` (Block): Network target for this slot.
  - `address` (String, required): Hostname or IP.
  - `port` (Number, required): gRPC port.
  - `transport` (String, optional): Transport name. Defaults to `grpc` when empty.
  - `tls` (Bool, optional): Reserved for TLS support in this block.

### Optional

- `name` (String): Human-readable name. If omitted, resource defaults to `device` for ID creation.
- `override_param_values_on_update` (Bool):
  - `false` (default behavior): parameters are applied on create only.
  - `true`: re-applies configured `parameters` on update.
- `parameters` (Dynamic): Parameter payload for the slot.
  - Supports either an object or a list/tuple of objects.
  - Objects are merged when a list is provided.
- `startup_command` (Single block): Commands to execute after create.
- `shutdown_command` (Single block): Commands to execute during destroy.

## startup_command and shutdown_command Block

Both blocks share the same schema:

- `commands` (List(String), required): Command OIDs executed in order.
- `values` (Dynamic, optional): One value per command. Use `null` for no-value commands.
- `status_foid` (String, optional): Parameter OID to poll after command execution.
- `status_success_value` (String, optional): Target value used by comparator.
- `status_success_comparator` (String, optional): One of `eq`, `ne`, `gt`, `lt`, `ge`, `le`.
- `timeout_seconds` (Number, optional): Poll timeout in seconds (defaults to 30).

Behavior notes:

- If a block is omitted, no commands are executed for that lifecycle phase.
- `shutdown_command` errors are reported as warnings so destroy can continue.

## Attributes Reference

Computed attributes:

- `id`: Resource identifier in the form `catena-<name>`.
- `parameters_out`: Writable parameters map for the slot.
- `full_parameters_out`: Full parameters map (including read-only) for the slot.
- `commands_out`: Command map for the slot.

All output maps are `map(string)` values from the device snapshot.

## Lifecycle Behavior

Create:

1. Configures client endpoint from `network`.
2. Applies `parameters` (if provided).
3. Runs `startup_command` (if provided).
4. Reads snapshot and populates computed outputs.

Read:

1. Refreshes snapshot outputs.
2. Validates `parameters` shape if present.

Update:

1. Reapplies `parameters` only when `override_param_values_on_update = true`.
2. Refreshes snapshot outputs.

Delete:

1. Runs `shutdown_command` if present.
2. Removes resource from state.

## Notes On Parameters

- `parameters` accepts nested values including numbers, strings, booleans, arrays, and objects.
- Arrays must be homogeneous; mixed-type arrays are rejected.
- Unknown or null dynamic values are treated as empty values where appropriate.

## Troubleshooting

If operations fail:

- Confirm `network.address` and `network.port` point to a reachable gRPC endpoint.
- Confirm parameter/command OIDs exist for the selected `slot`.
- Confirm parameter value shapes match device descriptors (especially arrays and nested structs).
- For status polling issues, verify `status_foid`, comparator, and expected value.
