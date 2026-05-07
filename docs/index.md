# ST2138 Provider

This provider manages Catena/ST2138 devices over gRPC.

The current implementation in this repository is centered on one primary resource,
`st2138_device`, with dynamic parameter writes and optional command blocks for lifecycle actions.

## Current Provider Surface

Provider source used by examples:

```hcl
terraform {
  required_providers {
    st2138 = {
      source  = "rossvideo/st2138"
      version = "0.0.2"
    }
  }
}
```

Provider block:

```hcl
provider "st2138" {}
```

Optional provider arguments currently recognized:

- `endpoint`: service endpoint such as `host:port`
- `transport`: transport type, typically `grpc`
- `devices_dir`: base directory for device executables/dockerfiles
- `executables_dir`: alias for `devices_dir`

## Resource And Data Source

Implemented types currently exposed by the provider:

- `st2138_device` resource
- `st2138_device_params` data source

See [docs/resources/device.md](docs/resources/device.md) for the detailed resource reference.

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
  - `startup_command` block
  - `shutdown_command` block
- Computed outputs:
  - `parameters_out`
  - `full_parameters_out`
  - `commands_out`

`startup_command` and `shutdown_command` are optional. If omitted, lifecycle actions skip those command executions.

## Recommended Example

The most up-to-date end-to-end example is:

- [examples/catena-test/main.tf](examples/catena-test/main.tf)
- [examples/catena-test/README.md](examples/catena-test/README.md)

That example demonstrates:

- Dynamic parameter payloads (`parameters`)
- Startup and shutdown command execution
- Status polling comparator behavior
- Output decoding for parameter/command maps

## Runtime Notes

- Transport defaults to `grpc` when not provided in `network.transport`.
- The provider reads/writes values via gRPC and retries parameter set operations for transient failures.
- `override_param_values_on_update = false` means parameter values are applied on create and not force-reapplied on every update.

## Local Development

During local provider development (before registry release propagation), use the local
development workflow documented in the catena example:

- [examples/catena-test/dev.tfrc](examples/catena-test/dev.tfrc)
- [examples/catena-test/README.md](examples/catena-test/README.md)

## Repository Links

- Provider implementation: [internal/provider/provider.go](internal/provider/provider.go)
- Device resource implementation: [internal/services/device/resource.go](internal/services/device/resource.go)
- Examples index: [examples/README.md](examples/README.md)
