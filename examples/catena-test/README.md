# catena-test Example

This example demonstrates the current `st2138_device` workflow used in this repository.
It is intended for local development and end-to-end validation of parameter writes,
startup and shutdown commands, and output decoding.

## What This Example Covers

- Creates one resource: `st2138_device.one_of_everything`
- Connects over gRPC using the `network` block
- Applies many parameter types using `parameters` (dynamic object/list shape)
- Runs `startup_command` after create
- Runs `shutdown_command` on destroy
- Exposes writable parameter values through `parameters_out`

## Configuration Walkthrough

The example in [main.tf](main.tf) uses:

- `slot = 0`: target slot for all gRPC requests
- `network`:
  - `address = "localhost"`
  - `port = 6254`
  - `transport = "grpc"`
  - `tls = false`
- `override_param_values_on_update = false`:
  - Parameters are applied on create
  - Updates refresh state without reapplying parameter values
- `parameters`:
  - Passed as a list containing one object
  - Includes scalar, array, struct, and nested struct examples
- `startup_command`:
  - Executes `/fib_start`
  - Polls `number_example` until comparator `ne` with value `0`
- `shutdown_command`:
  - Executes `/fib_stop` and `/fib_set`
  - Sends values `[null, { int32_value = 0 }]`
  - Polls `number_example` until comparator `eq` with value `0`

## Output Behavior

The output `device_params` iterates over `parameters_out` and attempts `jsondecode`.
This gives native OpenTofu values (numbers, lists, maps) when the provider returns
JSON-encoded strings, and falls back to the original string otherwise.

## Running The Example

From this directory:

```bash
tofu init
tofu plan
tofu apply
```

Destroy when finished:

```bash
tofu destroy
```

## Local Development Note

If registry publishing is not available yet, initialize with your local dev setup
(for example, using your local CLI config override and/or local plugin mirror).
Keep the `required_providers` source as `rossvideo/st2138` so the config remains
compatible with registry publishing later.
