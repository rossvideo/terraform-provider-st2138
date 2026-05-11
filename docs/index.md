# ST2138 Provider

This provider manages Catena/ST2138 devices over gRPC.
https://github.com/SMPTE/st2138-a
The current implementation in this repository is centered on one primary resource,
`st2138_device`, with dynamic parameter writes and optional command blocks for lifecycle actions.

## Current Provider Surface

Provider source used by examples:

```hcl
terraform {
  required_providers {
    st2138 = {
      source  = "rossvideo/st2138"
      version = "0.1.0"
    }
  }
}
```

Provider block:

```hcl
provider "st2138" {}
```


## Resource And Data Source

Implemented types currently exposed by the provider:

- `st2138_command` resource
- `st2138_device` resource
- `st2138_parameters` resource
- `st2138_device_params` data source

See [docs/resources/device.md](docs/resources/device.md) for the detailed resource reference.

Resource references:

- [docs/resources/command.md](docs/resources/command.md)
- [docs/resources/device.md](docs/resources/device.md)
- [docs/resources/parameters.md](docs/resources/parameters.md)

## st2138_device Highlights

The current `st2138_device` schema supports:

- Required:
  - `slot`
  - `network` block (`address`, `port`)
- Optional:
  - `name`
  - `override_param_values_on_update`
  - `parameters` (dynamic object or list-of-objects)
  - `network.transport`, `network.tls`
  - `startup_commands` block
  - `shutdown_commands` block
- Computed outputs:
  - `parameters_out`
  - `full_parameters_out`
  - `commands_out`
  - `status_value`

`startup_commands` and `shutdown_commands` are optional. If omitted, lifecycle actions skip those command executions.

## Recommended Example

The most up-to-date end-to-end example is:

- [examples/catena-test/main.tf](https://github.com/rossvideo/terraform-provider-st2138/blob/main/examples/catena-test/main.tf)

That example demonstrates:

- Dynamic parameter payloads (`parameters`)
- Startup and shutdown command execution
- Status polling comparator behavior
- Output decoding for parameter/command maps

## Runtime Notes

- Transport defaults to `grpc` when not provided in `network.transport`.
- The provider reads/writes values via gRPC and retries parameter set operations for transient failures.
- `override_param_values_on_update = false` means parameter values are applied on create and not force-reapplied on every update.

