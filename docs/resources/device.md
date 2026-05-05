# st2138_device Resource

Manages a device that implements the SMPTE ST2138 protocol for parameter control via gRPC.

This resource can manage both local Docker-based devices and remote gRPC devices, allowing you to configure device parameters, execute commands, and monitor device status.

## Example Usage

### Basic Device Configuration

```hcl
resource "st2138_device" "example" {
  name   = "my-device"
  slot   = 0
  
  params_map = {
    "/system/name" = "Production Device"
    "/inputs/0/enabled" = "true"
  }
}
```

### Remote gRPC Device

```hcl
resource "st2138_device" "remote" {
  name        = "remote-device"
  device_type = "remote-grpc"
  address     = "192.168.1.100"
  port        = 6254
  slot        = 0
  
  params_map = {
    "/audio/volume" = "75"
    "/video/format" = "1080p60"
  }
}
```

### Device with Start/Stop Commands

```hcl
resource "st2138_device" "managed" {
  name          = "managed-device"
  slot          = 0
  start_command = "/app/startup.sh"
  stop_command  = "/app/shutdown.sh"
  
  # Monitor device readiness
  device_status {
    oid         = "/system/status"
    ready_value = "running"
  }
  
  params {
    oid   = "/config/mode"
    value = "production"
  }
  
  params {
    oid   = "/config/debug"
    value = "false"
  }
}
```

### Device with Multiple Parameters

```hcl
resource "st2138_device" "video_processor" {
  name   = "video-proc-1"
  slot   = 0
  
  # Use params_map for bulk configuration
  params_map = {
    "/inputs/0/name"      = "Camera 1"
    "/inputs/0/enabled"   = "true"
    "/inputs/1/name"      = "Camera 2"
    "/inputs/1/enabled"   = "false"
    "/outputs/0/bitrate"  = "10000000"
    "/processing/quality" = "high"
  }
  
  # Apply all params even if unchanged
  apply_all = true
}
```

### Secure Remote Device (TLS)

```hcl
resource "st2138_device" "secure" {
  name        = "secure-device"
  device_type = "remote-grpc"
  address     = "grpcs://secure.example.com"  # grpcs:// enables TLS
  port        = 6254
  slot        = 0
  
  params_map = {
    "/security/encryption" = "enabled"
  }
}
```

## Argument Reference

### Required Arguments

* `slot` - (Required, Integer) Device slot ID used in gRPC calls. Must be >= 0.

### Optional Arguments

* `name` - (Optional, String) Human-readable device name. Used for identification and logging.
* `device_type` - (Optional, String) Device type identifier. Use `"remote-grpc"` for remote devices. Defaults to local device.
* `address` - (Optional, String) Remote device address (required when `device_type = "remote-grpc"`). Can be:
  - Hostname: `"192.168.1.100"`
  - URL with scheme: `"grpc://device.local"` or `"grpcs://secure.device.local"` (grpcs enables TLS)
* `port` - (Optional, Integer) Remote device port (required when `device_type = "remote-grpc"`). Typically 6254.
* `params_map` - (Optional, Map of String) Map of OID → value pairs for bulk parameter configuration. Values are parsed automatically:
  - Numbers: `"123"`, `"45.67"` → sent as numeric values
  - Booleans: `"true"`, `"false"` → sent as boolean values
  - Strings: `"any text"` → sent as string values
* `start_command` - (Optional, String) Command to execute on the device after parameter configuration. The resource will:
  1. Wait 3 seconds
  2. Execute the command via gRPC
  3. Wait for device to reach ready state (if `device_status` configured)
* `stop_command` - (Optional, String) Command to execute before device deletion. The resource will:
  1. Execute the command
  2. Wait 5 seconds
  3. Wait for device to leave ready state (if `device_status` configured)
* `apply_all` - (Optional, Boolean) When `true`, always sends all parameters on update, even if unchanged. Useful for devices that reset to defaults. Defaults to `false`.

### Blocks

#### `params` Block

Optional repeatable block for individual parameter configuration. Parameters set via `params` blocks override those in `params_map`.

* `oid` - (Required, String) Fully-qualified object identifier (e.g., `/inputs/0/name`).
* `value` - (Required, String) Value to set. Automatically parsed as string, number, or boolean.

Example:
```hcl
params {
  oid   = "/audio/input/0/gain"
  value = "12.5"
}

params {
  oid   = "/audio/input/0/muted"
  value = "false"
}
```

#### `device_status` Block

Optional single block for monitoring device readiness.

* `oid` - (Optional, String) Status OID to poll (e.g., `/system/state`).
* `endpoint` - (Optional, String) Deprecated alias for `oid`. Use `oid` instead.
* `ready_value` - (Optional, String) Expected value indicating device is ready (e.g., `"running"`, `"active"`).

When configured with `start_command`, the resource waits up to 60 seconds for the status OID to match `ready_value`.
When configured with `stop_command`, the resource waits up to 60 seconds for the status OID to differ from `ready_value`.

Example:
```hcl
device_status {
  oid         = "/system/operational_state"
  ready_value = "operational"
}
```

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

* `id` - Unique identifier for the device (format: `catena-<name>`).
* `container_id` - Docker container ID (for local Docker-based devices). Empty for remote devices.
* `status_value` - Current value of the `device_status` OID, refreshed on each read.

## Import

Devices can be imported using their name:

```bash
terraform import st2138_device.example my-device
```

## Behavior Notes

### Parameter Handling

- **Type Detection**: Values in `params_map` and `params` are automatically parsed:
  - `"123"` → integer
  - `"45.67"` → float
  - `"true"`, `"false"` → boolean
  - `"any other text"` → string
  
- **Update Behavior**: By default, only changed parameters are sent on update. Set `apply_all = true` to always send all parameters.

- **Precedence**: If the same OID appears in both `params_map` and `params` blocks, the `params` block value takes precedence.

### Device Lifecycle

1. **Create**: Sets all parameters, executes start command (if configured), waits for ready state
2. **Read**: Polls status OID to refresh `status_value`
3. **Update**: Sends changed parameters (or all if `apply_all = true`)
4. **Delete**: Executes stop command (if configured), waits for device to stop

### Retry Logic

Parameter setting includes automatic retry logic (3 attempts with exponential backoff) to handle transient connection issues during device startup.

### Remote vs Local Devices

**Local Devices** (default):
- Assumes device runs in Docker container
- Automatically discovers container ID
- Maps to Docker host networking

**Remote Devices** (`device_type = "remote-grpc"`):
- Connects directly to remote gRPC endpoint
- Requires `address` and `port`
- Supports TLS via `grpcs://` scheme
- No Docker container management

## Troubleshooting

### Connection Errors

If you see connection errors, verify:
- Device is running and accessible
- Port is correct (typically 6254)
- Firewall allows gRPC traffic
- For remote devices: address and port are correct
- For local devices: Docker container is running

### Parameter Set Failures

If parameter setting fails:
- Verify OID path is correct (use gRPC inspection tools)
- Check value format matches parameter type
- Ensure device is in ready state before setting parameters
- Try setting `apply_all = true` if device resets parameters unexpectedly

### Status Polling Timeouts

If `WaitReady` times out (60s default):
- Verify `device_status.oid` is correct
- Check `ready_value` matches expected device state
- Inspect device logs for startup issues
- Increase timeout by adjusting `start_command` timing
