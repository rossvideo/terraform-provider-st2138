# ST2138 Provider

The ST2138 provider enables Terraform to manage devices that implement the SMPTE ST2138 protocol for parameter control via gRPC.

## Example Usage

```hcl
terraform {
  required_providers {
    st2138 = {
      source  = "rossvideo/st2138"
      version = "~> 1.0"
    }
  }
}

provider "st2138" {}

resource "st2138_device" "example" {
  name = "my-device"
  slot = 0
  
  params_map = {
    "/system/name" = "Production Device"
    "/config/mode" = "live"
  }
}
```

## Provider Configuration

The ST2138 provider can be configured with no arguments for default behavior, or with optional settings for customization.

### Basic Configuration

```hcl
provider "st2138" {}
```

This uses default settings:
- gRPC transport
- localhost:6254 as default endpoint
- No TLS

### Custom Configuration

```hcl
provider "st2138" {
  # Provider configuration options would go here
  # (Currently no provider-level configuration required)
}
```

## Authentication

The ST2138 provider uses gRPC for communication. Authentication is handled at the network/transport layer:

- **Local devices**: No authentication required (localhost connections)
- **Remote devices**: Configure authentication via:
  - Network security (VPN, private networks)
  - gRPC TLS (`grpcs://` scheme)
  - Firewall rules

## Resources

### st2138_device

Manages a device implementing the SMPTE ST2138 protocol.

**Key features:**
- Configure device parameters via gRPC
- Support for local and remote devices
- Lifecycle management (start/stop commands)
- Status monitoring and health checks
- Bulk parameter configuration
- Automatic type detection for values

See [resources/device.md](./resources/device.md) for complete documentation.

## gRPC Transport

The provider communicates with devices using gRPC (Google Remote Procedure Call):

- **Default port**: 6254
- **Protocol**: HTTP/2
- **TLS support**: Use `grpcs://` scheme in address
- **Automatic reconnection**: Built-in retry logic for transient failures

## OID Paths

Device parameters are identified by OID (Object Identifier) paths:

- Format: `/path/to/parameter`
- Examples:
  - `/system/name` - System name
  - `/inputs/0/enabled` - First input enable status
  - `/config/mode` - Configuration mode
  
OID paths are device-specific. Consult your device documentation for available parameters.

## Environment Variables

The provider does not currently use environment variables for configuration. All settings are specified in the provider and resource blocks.

## Debugging

To enable debug logging for the provider:

```bash
export TF_LOG=DEBUG
terraform apply
```

This shows detailed gRPC communication and parameter setting operations.

## Compatibility

- **Terraform**: >= 0.13
- **Devices**: Any device implementing SMPTE ST2138 gRPC protocol
- **Platforms**: Linux, macOS, Windows

## Known Limitations

1. **Read-only parameters**: Some device parameters may be read-only and cannot be set via the provider
2. **Complex types**: Currently supports string, number, and boolean values. Complex nested structures may require multiple parameters
3. **Concurrent access**: While the provider includes retry logic, extremely high concurrency may require rate limiting

## Migration from Older Versions

This provider is designed for new installations. If migrating from custom scripts:

1. Inventory existing device configurations
2. Map parameter names to OID paths
3. Create Terraform configurations using `params_map`
4. Import existing devices if possible
5. Test in non-production environment first

## Support and Contributing

- **Issues**: Report bugs and feature requests on GitHub
- **Documentation**: See `../examples/` for usage examples
- **Contributing**: Pull requests welcome

## License

This provider is available under the Mozilla Public License 2.0.
