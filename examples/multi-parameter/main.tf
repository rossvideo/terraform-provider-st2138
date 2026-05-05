# Multi-Parameter Configuration Example

This example shows advanced parameter management techniques.

## Usage

```bash
terraform init
terraform plan
terraform apply
```

## Configuration

```hcl
terraform {
  required_providers {
    st2138 = {
      source = "rossvideo/st2138"
    }
  }
}

provider "st2138" {}

# Example: Video processor with extensive configuration
resource "st2138_device" "video_processor" {
  name = "video-proc-1"
  slot = 0

  # Use params_map for bulk configuration
  # This is ideal for large numbers of related parameters
  params_map = {
    # System configuration
    "/system/name"        = "Video Processor 1"
    "/system/location"    = "Rack 2, Unit 5"
    "/system/description" = "Main video processing unit"

    # Input configuration (4 inputs)
    "/inputs/0/name"    = "Camera 1 - Studio A"
    "/inputs/0/enabled" = "true"
    "/inputs/0/format"  = "1080p60"
    "/inputs/1/name"    = "Camera 2 - Studio B"
    "/inputs/1/enabled" = "true"
    "/inputs/1/format"  = "1080p60"
    "/inputs/2/name"    = "Graphics Feed"
    "/inputs/2/enabled" = "true"
    "/inputs/2/format"  = "1080p60"
    "/inputs/3/name"    = "Backup"
    "/inputs/3/enabled" = "false"
    "/inputs/3/format"  = "720p60"

    # Output configuration
    "/outputs/0/name"      = "Program Out"
    "/outputs/0/enabled"   = "true"
    "/outputs/0/bitrate"   = "10000000"
    "/outputs/0/format"    = "1080p60"
    "/outputs/1/name"      = "Preview Out"
    "/outputs/1/enabled"   = "true"
    "/outputs/1/bitrate"   = "5000000"
    "/outputs/1/format"    = "720p60"

    # Processing settings
    "/processing/quality"      = "high"
    "/processing/denoise"      = "true"
    "/processing/color_space"  = "rec709"
    "/processing/gpu_accel"    = "true"
  }

  # Use params blocks for computed or conditional values
  # These override params_map if there's overlap
  params {
    oid   = "/processing/threads"
    value = "8"  # Could be computed from available CPU
  }

  params {
    oid   = "/outputs/0/quality"
    value = "95"  # High quality for program output
  }

  params {
    oid   = "/outputs/1/quality"
    value = "85"  # Lower quality acceptable for preview
  }

  # Apply all parameters on every update
  # Useful if device resets parameters unexpectedly
  apply_all = true
}

# Example: Using locals for repeated parameter patterns
locals {
  audio_channels = {
    for i in range(8) : i => {
      name    = "Channel ${i + 1}"
      enabled = i < 4 ? "true" : "false"  # First 4 channels enabled
      gain    = "0.0"
      muted   = "false"
    }
  }

  # Flatten audio channel config into params_map format
  audio_params = merge([
    for ch, config in local.audio_channels : {
      "/audio/inputs/${ch}/name"    = config.name
      "/audio/inputs/${ch}/enabled" = config.enabled
      "/audio/inputs/${ch}/gain"    = config.gain
      "/audio/inputs/${ch}/muted"   = config.muted
    }
  ]...)
}

resource "st2138_device" "audio_processor" {
  name = "audio-proc-1"
  slot = 1

  # Combine computed audio params with static config
  params_map = merge(
    local.audio_params,
    {
      "/system/name"          = "Audio Processor 1"
      "/processing/mode"      = "stereo"
      "/processing/sample_rate" = "48000"
      "/outputs/0/format"     = "pcm"
    }
  )
}

# Example: Conditional parameter configuration
variable "environment" {
  type    = string
  default = "production"
}

resource "st2138_device" "conditional" {
  name = "conditional-device"
  slot = 2

  params_map = merge(
    {
      "/system/name" = "Conditional Config Device"
      "/config/env"  = var.environment
    },
    # Production-specific settings
    var.environment == "production" ? {
      "/logging/level"       = "warn"
      "/monitoring/enabled"  = "true"
      "/debug/verbose"       = "false"
      "/performance/optimize" = "true"
    } : {},
    # Development-specific settings
    var.environment == "development" ? {
      "/logging/level"       = "debug"
      "/monitoring/enabled"  = "false"
      "/debug/verbose"       = "true"
      "/performance/optimize" = "false"
    } : {}
  )
}

# Outputs
output "video_processor_id" {
  value = st2138_device.video_processor.id
}

output "audio_processor_params" {
  value       = local.audio_params
  description = "Computed audio channel parameters"
}

output "environment_config" {
  value = {
    environment = var.environment
    device_id   = st2138_device.conditional.id
  }
}
```

## Techniques Demonstrated

### 1. Bulk Parameter Setting

Use `params_map` for large numbers of related parameters. This is more readable and maintainable than individual `params` blocks.

### 2. Parameter Precedence

`params` blocks override `params_map` entries. Use this for:
- Computed values
- Conditional settings
- Values that depend on other resources

### 3. Computed Configuration

Use `locals` and `for` expressions to generate repetitive parameter patterns.

### 4. Conditional Configuration

Use `merge()` with conditional expressions to apply different settings based on variables.

### 5. apply_all Flag

Set `apply_all = true` when:
- Device resets parameters on restart
- You want guaranteed state on every apply
- Troubleshooting parameter sync issues

## Type Handling

All values are automatically parsed:
- `"123"` → integer
- `"45.67"` → float
- `"true"`, `"false"` → boolean
- `"text"` → string

## Performance Notes

- Bulk parameter setting via `params_map` is efficient
- Parameters are sent sequentially to ensure ordering
- Retry logic handles transient failures
- Use `apply_all = false` (default) to minimize gRPC calls on update
