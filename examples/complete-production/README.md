# Complete Production Example

This example demonstrates a realistic production deployment with multiple devices, proper lifecycle management, and comprehensive monitoring.

## Architecture

```
┌─────────────────────┐
│  Video Processor    │ Slot 0 - Primary video processing
│  (Primary)          │ 
└─────────────────────┘
          │
          ├─────────────────────┐
          │                     │
┌─────────────────────┐ ┌─────────────────────┐
│  Audio Processor    │ │  Graphics Engine    │
│  (Slot 1)           │ │  (Remote)           │
└─────────────────────┘ └─────────────────────┘
```

## Usage

1. Update variables in `terraform.tfvars`:
```hcl
environment        = "production"
graphics_host      = "192.168.1.50"
enable_monitoring  = true
```

2. Deploy:
```bash
terraform init
terraform plan
terraform apply
```

3. Monitor:
```bash
terraform output device_status
```

## Configuration Files

### main.tf

```hcl
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    st2138 = {
      source  = "rossvideo/st2138"
      version = "~> 1.0"
    }
  }
}

provider "st2138" {}

# Primary video processor - handles main video routing
resource "st2138_device" "video_primary" {
  name = "video-proc-primary"
  slot = 0

  # Lifecycle management
  start_command = "/app/scripts/start-video-proc.sh"
  stop_command  = "/app/scripts/stop-video-proc.sh"

  device_status {
    oid         = "/system/operational_state"
    ready_value = "operational"
  }

  # System configuration
  params_map = merge(
    local.common_params,
    {
      "/system/name"        = "Primary Video Processor"
      "/system/role"        = "primary"
      "/system/location"    = var.rack_location
      
      # Video inputs
      "/inputs/video/0/name"    = "Camera 1 - Main"
      "/inputs/video/0/enabled" = "true"
      "/inputs/video/0/format"  = var.video_format
      "/inputs/video/1/name"    = "Camera 2 - Backup"
      "/inputs/video/1/enabled" = "true"
      "/inputs/video/1/format"  = var.video_format
      "/inputs/video/2/name"    = "Graphics Feed"
      "/inputs/video/2/enabled" = "true"
      "/inputs/video/2/format"  = var.video_format
      
      # Video outputs
      "/outputs/video/0/name"    = "Program Output"
      "/outputs/video/0/enabled" = "true"
      "/outputs/video/0/bitrate" = var.video_bitrate
      "/outputs/video/1/name"    = "Preview Output"
      "/outputs/video/1/enabled" = "true"
      "/outputs/video/1/bitrate" = tostring(var.video_bitrate / 2)
      
      # Processing
      "/processing/mode"         = "live"
      "/processing/latency"      = "low"
      "/processing/quality"      = var.processing_quality
      "/processing/color_space"  = "rec709"
    }
  )

  # Always apply all parameters to ensure consistency
  apply_all = var.apply_all_params
}

# Audio processor - handles audio routing and mixing
resource "st2138_device" "audio_primary" {
  name = "audio-proc-primary"
  slot = 1

  start_command = "/app/scripts/start-audio-proc.sh"
  stop_command  = "/app/scripts/stop-audio-proc.sh"

  device_status {
    oid         = "/system/state"
    ready_value = "running"
  }

  params_map = merge(
    local.common_params,
    local.audio_channel_params,
    {
      "/system/name"     = "Primary Audio Processor"
      "/system/role"     = "primary"
      
      # Audio configuration
      "/audio/sample_rate"    = "48000"
      "/audio/bit_depth"      = "24"
      "/audio/mode"           = "stereo"
      
      # Outputs
      "/outputs/audio/0/name"    = "Program Audio"
      "/outputs/audio/0/enabled" = "true"
      "/outputs/audio/0/format"  = "pcm"
    }
  )

  # Audio processor depends on video processor
  depends_on = [st2138_device.video_primary]
}

# Remote graphics engine
resource "st2138_device" "graphics_remote" {
  count = var.enable_graphics ? 1 : 0

  name        = "graphics-engine"
  device_type = "remote-grpc"
  address     = var.graphics_use_tls ? "grpcs://${var.graphics_host}" : var.graphics_host
  port        = var.graphics_port
  slot        = 0  # Slot on remote device

  params_map = {
    "/system/name"          = "Graphics Engine"
    "/graphics/template"    = var.graphics_template
    "/graphics/resolution"  = var.video_format
    "/graphics/refresh_rate" = "60"
    "/outputs/0/enabled"    = "true"
    "/sync/source"          = st2138_device.video_primary.id
  }

  device_status {
    oid         = "/system/status"
    ready_value = "online"
  }
}

# Monitoring device (read-only, just for status)
resource "st2138_device" "monitor" {
  count = var.enable_monitoring ? 1 : 0

  name = "system-monitor"
  slot = 3

  params_map = {
    "/system/name"        = "Production Monitor"
    "/monitor/video_src"  = st2138_device.video_primary.id
    "/monitor/audio_src"  = st2138_device.audio_primary.id
    "/monitor/alerts"     = "enabled"
    "/monitor/log_level"  = var.environment == "production" ? "warn" : "debug"
  }
}
```

### variables.tf

```hcl
variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "production"
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "rack_location" {
  description = "Physical rack location"
  type        = string
  default     = "Rack 1, Unit 3-5"
}

variable "video_format" {
  description = "Video format (resolution and frame rate)"
  type        = string
  default     = "1080p60"
}

variable "video_bitrate" {
  description = "Video output bitrate in bps"
  type        = number
  default     = 10000000
}

variable "processing_quality" {
  description = "Processing quality level"
  type        = string
  default     = "high"
  validation {
    condition     = contains(["low", "medium", "high", "ultra"], var.processing_quality)
    error_message = "Quality must be low, medium, high, or ultra."
  }
}

variable "enable_graphics" {
  description = "Enable graphics engine device"
  type        = bool
  default     = true
}

variable "graphics_host" {
  description = "Graphics engine hostname or IP"
  type        = string
  default     = "localhost"
}

variable "graphics_port" {
  description = "Graphics engine gRPC port"
  type        = number
  default     = 6254
}

variable "graphics_use_tls" {
  description = "Use TLS for graphics engine connection"
  type        = bool
  default     = false
}

variable "graphics_template" {
  description = "Graphics template name"
  type        = string
  default     = "standard_lower_third"
}

variable "enable_monitoring" {
  description = "Enable monitoring device"
  type        = bool
  default     = true
}

variable "apply_all_params" {
  description = "Always apply all parameters on update"
  type        = bool
  default     = false
}

variable "audio_channels" {
  description = "Number of audio channels to configure"
  type        = number
  default     = 8
  validation {
    condition     = var.audio_channels >= 2 && var.audio_channels <= 16
    error_message = "Audio channels must be between 2 and 16."
  }
}
```

### locals.tf

```hcl
locals {
  # Common parameters applied to all devices
  common_params = {
    "/system/environment"    = var.environment
    "/system/deployed_by"    = "Terraform"
    "/system/deployed_at"    = timestamp()
    "/monitoring/enabled"    = var.enable_monitoring ? "true" : "false"
    "/logging/level"         = var.environment == "production" ? "info" : "debug"
    "/logging/destination"   = "syslog"
  }

  # Generate audio channel configuration
  audio_channel_params = merge([
    for i in range(var.audio_channels) : {
      "/audio/inputs/${i}/name"    = "Audio Channel ${i + 1}"
      "/audio/inputs/${i}/enabled" = i < 4 ? "true" : "false"
      "/audio/inputs/${i}/gain"    = "0.0"
      "/audio/inputs/${i}/muted"   = "false"
    }
  ]...)
}
```

### outputs.tf

```hcl
output "device_status" {
  description = "Status of all devices"
  value = {
    video_primary = {
      id     = st2138_device.video_primary.id
      status = st2138_device.video_primary.status_value
    }
    audio_primary = {
      id     = st2138_device.audio_primary.id
      status = st2138_device.audio_primary.status_value
    }
    graphics = var.enable_graphics ? {
      id      = st2138_device.graphics_remote[0].id
      status  = st2138_device.graphics_remote[0].status_value
      endpoint = "${st2138_device.graphics_remote[0].address}:${st2138_device.graphics_remote[0].port}"
    } : null
  }
}

output "video_config" {
  description = "Video configuration summary"
  value = {
    format     = var.video_format
    bitrate    = var.video_bitrate
    quality    = var.processing_quality
    device_id  = st2138_device.video_primary.id
  }
}

output "deployment_info" {
  description = "Deployment information"
  value = {
    environment   = var.environment
    location      = var.rack_location
    deployed_at   = local.common_params["/system/deployed_at"]
    graphics_enabled = var.enable_graphics
    monitoring_enabled = var.enable_monitoring
  }
}
```

### terraform.tfvars.example

```hcl
# Copy to terraform.tfvars and customize

environment        = "production"
rack_location      = "Main Facility - Rack 2"
video_format       = "1080p60"
video_bitrate      = 10000000
processing_quality = "high"

# Graphics engine
enable_graphics    = true
graphics_host      = "192.168.1.50"
graphics_port      = 6254
graphics_use_tls   = false
graphics_template  = "lower_third_v2"

# Monitoring
enable_monitoring  = true

# Audio
audio_channels     = 8

# Update behavior
apply_all_params   = false
```

## Deployment Workflow

### Initial Deployment

```bash
# 1. Copy and customize variables
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars

# 2. Initialize
terraform init

# 3. Plan and review
terraform plan -out=deployment.plan

# 4. Apply
terraform apply deployment.plan
```

### Making Changes

```bash
# Update configuration
vim main.tf

# Preview changes
terraform plan

# Apply changes
terraform apply
```

### Disaster Recovery

```bash
# Recreate all devices
terraform destroy
terraform apply

# Recreate single device
terraform taint st2138_device.video_primary
terraform apply
```

## Monitoring

### Check Device Status

```bash
terraform output device_status
```

### Verify Configuration

```bash
terraform show
terraform state list
terraform state show st2138_device.video_primary
```

## Production Checklist

- [ ] All variables configured in `terraform.tfvars`
- [ ] Graphics host reachable on port 6254
- [ ] Start/stop scripts exist and are executable
- [ ] Device status OIDs correct for your devices
- [ ] Backup of terraform state configured
- [ ] Monitoring enabled
- [ ] Tested in staging environment
- [ ] Rollback plan documented

## Troubleshooting

### Devices fail to start

Check logs:
```bash
terraform apply -refresh-only
terraform output device_status
```

### Parameter changes don't apply

Try forcing re-application:
```bash
terraform apply -var="apply_all_params=true"
```

### Graphics engine unreachable

Verify connectivity:
```bash
nc -zv 192.168.1.50 6254
```

## Notes

- Device startup is sequential (audio waits for video)
- Graphics engine is optional (controlled by `enable_graphics`)
- Monitoring device tracks all other devices
- All timestamps are in UTC
- State file contains sensitive configuration
