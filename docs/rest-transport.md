# REST/HTTP Transport Support

## Overview

The Terraform Provider for ST2138 now supports both **gRPC** and **REST/HTTP** transports for communicating with SMPTE ST2138-A compliant Catena devices. This document describes the REST transport implementation and usage.

## OpenAPI Specification

The REST API is fully defined by the SMPTE ST2138-A OpenAPI specification:
- **Repository**: https://github.com/SMPTE/st2138-a
- **OpenAPI Spec**: https://github.com/SMPTE/st2138-a/blob/main/interface/openapi/openapi.yaml

## Supported Endpoints

The REST transport implementation supports the following API endpoints:

### Device Operations
- `GET /{slot}/stream` - Stream device information
- `GET /{slot}/value/{fqoid}` - Get parameter value
- `PUT /{slot}/value/{fqoid}` - Set parameter value
- `GET /{slot}/param-info/{fqoid}` - Get parameter descriptor

### Command Execution
- `POST /{slot}/command/{fqoid}/stream` - Execute command with streaming response

### Base URL
- **Default**: `https://device.catenamedia.tv:443/st2138-api/v1` (from OpenAPI spec)
- **Configurable**: Derived from provider endpoint configuration

## Configuration

### Using HTTP Transport

```hcl
terraform {
  required_providers {
    st2138 = {
      source = "registry.opentofu.org/rossvideo/st2138"
    }
  }
}

provider "st2138" {
  endpoint  = "http://device.example.com:8080"
  transport = "http"
}
```

### Using HTTPS Transport

```hcl
provider "st2138" {
  endpoint  = "https://device.example.com:443"
  transport = "https"
}
```

### Using REST Auto-detect

```hcl
provider "st2138" {
  endpoint  = "https://device.example.com:443"
  transport = "rest"  # Auto-detects https:// scheme
}
```

## Implementation Details

### HTTP Client Architecture

The HTTP transport is implemented in [internal/client/http.go](../internal/client/http.go) with the following components:

#### HTTPClient Type
- Manages HTTP connections and request/response handling
- Supports TLS for HTTPS connections
- Provides timeout and deadline management
- Compatible with OpenAPI specification endpoints

#### HTTPStream Type
- Mimics gRPC streaming interfaces for compatibility
- Provides `Recv()` and `RecvCommandResponse()` methods
- Automatically manages JSON decoding of responses

### Transport Dispatch

The `Client` type in [internal/client/client.go](../internal/client/client.go) implements transport detection:

- Checks `Transport` field during method execution
- Routes to `httpClient` for HTTP/HTTPS/rest transports
- Routes to `rpcClient` for gRPC transport
- Maintains connection pools for both transports

### Key Methods Supporting Both Transports

#### Device Snapshot
```go
func (c *Client) GetDeviceSnapshot(ctx context.Context, slot uint32) (*DeviceSnapshot, error)
// GET /{slot}/stream via HTTP
// DeviceRequest RPC via gRPC
```

#### Parameter Operations
```go
func (c *Client) GetRawValue(ctx context.Context, slot uint32, oid string) (*st2138pb.Value, error)
// GET /{slot}/value/{oid} via HTTP
// GetValue RPC via gRPC

func (c *Client) SetStringValue(ctx context.Context, slot uint32, oid string, value string) error
// PUT /{slot}/value/{oid} via HTTP
// SetValue RPC via gRPC

func (c *Client) SetNumberValue(ctx context.Context, slot uint32, oid string, n float64) error
// PUT /{slot}/value/{oid} via HTTP
// SetValue RPC via gRPC
```

#### Command Execution
```go
func (c *Client) ExecuteCommand(ctx context.Context, slot uint32, oid string, value *st2138pb.Value) error
// POST /{slot}/command/{oid}/stream via HTTP
// ExecuteCommand RPC via gRPC

func (c *Client) RunStart(ctx context.Context, slot uint32, commandOID string) error
// HTTP: POST to command endpoint with empty value
// gRPC: ExecuteCommand with empty payload

func (c *Client) RunStop(ctx context.Context, slot uint32, commandOID string) error
// HTTP: POST to command endpoint with empty value
// gRPC: ExecuteCommand with empty payload
```

#### Status Monitoring
```go
func (c *Client) GetStringValue(ctx context.Context, slot uint32, oid string) (string, error)
// GET /{slot}/value/{oid} via HTTP
// GetValue RPC via gRPC

func (c *Client) WaitReady(ctx context.Context, slot uint32, endpoint string, readyValue string, timeout time.Duration) error
// Polling via GetStringValue (supports both transports)

func (c *Client) WaitNotReady(ctx context.Context, slot uint32, endpoint string, readyValue string, timeout time.Duration) error
// Polling via GetStringValue (supports both transports)
```

## OID Format

OIDs are normalized consistently across transports:

### Input Format
- Can be provided with or without leading slash
- Examples: `/device/status` or `device/status`

### HTTP Encoding
- Leading slash is added if missing
- Forward slashes are converted to colons in URL path
- Example: `/device/status` → `device:status` in URL path
- Full URL: `GET /st2138-api/v1/{slot}/value/device:status`

### gRPC Format
- Leading slash is required
- Full path in message: `/device/status`

## Error Handling

Both transports provide consistent error handling:

### HTTP Errors
- Returns HTTP status code and body content in error messages
- Example: `"device request failed with status 404: resource not found"`

### gRPC Errors
- Returns gRPC error status and details
- Automatically propagated as Go errors

## Connection Management

### HTTP
- Lazy connection establishment (on first request)
- Configurable timeouts (default 30 seconds)
- Automatic connection pooling and reuse
- TLS configuration based on scheme

### gRPC
- Lazy connection establishment (on first request)
- Connection caching per client instance
- Automatic reconnection on failure
- Explicit endpoint changes close and re-establish connections

## Testing

### Unit Tests
Located in [internal/client/](../internal/client/):
- `client_test.go` - Basic client operations
- `device_params.go` - Device parameter handling
- `params_errors_test.go` - Parameter error scenarios
- `device_errors_test.go` - Device operation error handling

### Integration Tests
The integration tests in `integration_test.go` verify:
- Connection establishment for both transports
- Parameter get/set operations
- Command execution (start/stop)
- Status polling and waiting

### Example Usage
See [examples/](../examples/) for complete Terraform configurations using both transports:
- `examples/basic-device/main.tf` - gRPC transport example
- `examples/catena-test/main.tf` - Docker-based test environment

## Compliance

The REST transport implementation follows the SMPTE ST2138-A OpenAPI specification:
- **Specification URL**: https://github.com/SMPTE/st2138-a/blob/main/interface/openapi/openapi.yaml
- **Base Path**: `/st2138-api/v1`
- **Security**: OAuth2 (configured per endpoint)
- **Response Format**: JSON with protobuf-equivalent structure

## Future Enhancements

Potential improvements for REST transport:
1. **WebSocket Streams**: For real-time parameter subscriptions
2. **OAuth2 Integration**: Automatic token refresh and management
3. **Response Compression**: gzip/brotli for large payloads
4. **Connection Pooling**: Per-host limit configuration
5. **Circuit Breaker**: Automatic failover on repeated failures
6. **Metrics Collection**: Prometheus-compatible export

## Troubleshooting

### Common Issues

**Issue**: `transport "http" is not grpc` error
- **Cause**: Old gRPC-only code path
- **Solution**: Ensure provider version is updated with HTTP support

**Issue**: HTTP connection times out
- **Cause**: Endpoint unreachable or slow network
- **Solution**: Check endpoint URL format and network connectivity

**Issue**: 404 errors on REST endpoints
- **Cause**: Invalid OID format or incorrect slot number
- **Solution**: Verify OID normalization and slot configuration

**Issue**: JSON decode errors
- **Cause**: Server returning non-JSON response
- **Solution**: Verify server is using REST API, not gRPC

## Related Documentation

- [README.md](../README.md) - Provider overview and configuration
- [DONE.md](../DONE.md) - Project completion status
- [TODO.md](../TODO.md) - Planned future improvements
- [ST2138-A OpenAPI Specification](https://github.com/SMPTE/st2138-a/blob/main/interface/openapi/openapi.yaml)
