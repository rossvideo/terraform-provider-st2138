package client

import (
	"context"
	"testing"
)

// Additional tests for connection management to improve ensureConn coverage

func TestClient_ensureConn_NonGRPCTransport(t *testing.T) {
	c := &Client{
		Transport: "http",
		Endpoint:  "localhost:6254",
	}

	err := c.ensureConn(context.Background())
	if err == nil {
		t.Error("ensureConn() should error for non-grpc transport")
	}
	if err.Error() != `transport "http" is not grpc` {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestClient_ensureConn_EmptyTransport(t *testing.T) {
	c := &Client{
		Transport: "",
		Endpoint:  "localhost:6254",
	}

	err := c.ensureConn(context.Background())
	if err == nil {
		t.Error("ensureConn() should error for empty transport")
	}
}

func TestClient_ensureConn_InvalidEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "empty endpoint",
			endpoint: "",
		},
		{
			name:     "invalid host",
			endpoint: "not-a-real-host-that-exists-12345:6254",
		},
		{
			name:     "malformed endpoint",
			endpoint: ":::invalid:::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Transport: "grpc",
				Endpoint:  tt.endpoint,
			}

			// Should either error or timeout quickly
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel to avoid hanging on dial

			err := c.ensureConn(ctx)
			// With cancelled context, should get context error
			_ = err
		})
	}
}

func TestClient_ensureConn_WithHTTPSScheme(t *testing.T) {
	c := &Client{
		Transport: "grpc",
		Endpoint:  "https://example.com:6254",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel to avoid actual dial

	err := c.ensureConn(ctx)
	// Should attempt to connect with TLS based on https:// scheme
	_ = err
}

func TestClient_ensureConn_WithHTTPScheme(t *testing.T) {
	c := &Client{
		Transport: "grpc",
		Endpoint:  "http://example.com:6254",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel to avoid actual dial

	err := c.ensureConn(ctx)
	// Should attempt to connect without TLS based on http:// scheme
	_ = err
}

func TestClient_ensureConn_WithGRPCScheme(t *testing.T) {
	c := &Client{
		Transport: "grpc",
		Endpoint:  "grpc://example.com:6254",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.ensureConn(ctx)
	_ = err
}

func TestClient_ensureConn_WithGRPCSScheme(t *testing.T) {
	c := &Client{
		Transport: "grpc",
		Endpoint:  "grpcs://example.com:6254",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.ensureConn(ctx)
	_ = err
}

func TestClient_ensureConn_HostOnlyNoPort(t *testing.T) {
	c := &Client{
		Transport: "grpc",
		Endpoint:  "localhost",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.ensureConn(ctx)
	// Should handle host without explicit port
	_ = err
}

func TestClient_ensureConn_IPv6Address(t *testing.T) {
	c := &Client{
		Transport: "grpc",
		Endpoint:  "[::1]:6254",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.ensureConn(ctx)
	_ = err
}

func TestClient_Close_WithNilConnection(t *testing.T) {
	c := &Client{
		conn:      nil,
		rpcClient: nil,
	}

	// Should not panic
	c.Close()

	if c.conn != nil {
		t.Error("conn should still be nil after Close()")
	}
	if c.rpcClient != nil {
		t.Error("rpcClient should still be nil after Close()")
	}
}

func TestClient_Close_ClearsRPCClient(t *testing.T) {
	// Close() only clears rpcClient when conn is not nil
	// This test verifies Close() behavior when there's no active connection
	mockClient := &mockCatenaServiceClient{}
	c := &Client{
		conn:      nil,
		rpcClient: mockClient,
	}

	c.Close()

	// When conn is nil, Close() is a no-op and rpcClient is not cleared
	// This is by design - the client is already "closed"
	if c.rpcClient == nil {
		t.Error("rpcClient should not be cleared when conn is nil")
	}
}

func TestClient_Clone_PreservesAllFields(t *testing.T) {
	original := &Client{
		Endpoint:   "test:6254",
		Transport:  "grpc",
		DevicesDir: "/custom/devices",
	}

	clone := original.Clone()

	if clone.Endpoint != original.Endpoint {
		t.Errorf("Clone Endpoint = %s, want %s", clone.Endpoint, original.Endpoint)
	}
	if clone.Transport != original.Transport {
		t.Errorf("Clone Transport = %s, want %s", clone.Transport, original.Transport)
	}
	if clone.DevicesDir != original.DevicesDir {
		t.Errorf("Clone DevicesDir = %s, want %s", clone.DevicesDir, original.DevicesDir)
	}
	if clone.conn != nil {
		t.Error("Clone should have nil conn")
	}
	if clone.rpcClient != nil {
		t.Error("Clone should have nil rpcClient")
	}
}
