package client

import (
	"context"
	"errors"
	"testing"
	"time"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
)

// Additional tests for device operations to improve coverage

func TestClient_WaitReady_ContextTimeout(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			// Always return not-ready to force timeout
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "starting"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	// Short timeout to force timeout error
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := c.WaitReady(ctx, 1, "/status", "ready", 100*time.Millisecond)
	if err == nil {
		t.Error("WaitReady() should timeout when value never becomes ready")
	}
}

func TestClient_WaitReady_GetValueError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("connection lost")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.WaitReady(context.Background(), 1, "/status", "ready", 1*time.Second)
	if err == nil {
		t.Error("WaitReady() should return error when GetValue fails")
	}
}

func TestClient_WaitReady_ImmediatelyReady(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			// Already ready on first call
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "ready"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	start := time.Now()
	err := c.WaitReady(context.Background(), 1, "/status", "ready", 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("WaitReady() error = %v", err)
	}

	// Should return immediately, not wait the full timeout
	if elapsed > 500*time.Millisecond {
		t.Errorf("WaitReady() took too long when already ready: %v", elapsed)
	}
}

func TestClient_WaitNotReady_ContextTimeout(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			// Always return ready to force timeout
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "ready"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := c.WaitNotReady(ctx, 1, "/status", "ready", 100*time.Millisecond)
	if err == nil {
		t.Error("WaitNotReady() should timeout when value stays ready")
	}
}

func TestClient_WaitNotReady_GetValueError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("device unreachable")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.WaitNotReady(context.Background(), 1, "/status", "ready", 1*time.Second)
	if err == nil {
		t.Error("WaitNotReady() should return error when GetValue fails")
	}
}

func TestClient_WaitNotReady_ImmediatelyNotReady(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			// Already not ready on first call
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "stopped"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	start := time.Now()
	err := c.WaitNotReady(context.Background(), 1, "/status", "ready", 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("WaitNotReady() error = %v", err)
	}

	// Should return immediately
	if elapsed > 500*time.Millisecond {
		t.Errorf("WaitNotReady() took too long when already not ready: %v", elapsed)
	}
}

func TestClient_RunStart_ExecuteError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		executeCommandFunc: func(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error) {
			return nil, errors.New("command failed")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.RunStart(context.Background(), 1, "/commands/start")
	if err == nil {
		t.Error("RunStart() should return error when ExecuteCommand fails")
	}
}

func TestClient_RunStart_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "http", // Invalid transport
		Endpoint:  "localhost:6254",
	}

	err := c.RunStart(context.Background(), 1, "/commands/start")
	if err == nil {
		t.Error("RunStart() should return error when ensureConn fails")
	}
}

func TestClient_RunStop_ExecuteError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		executeCommandFunc: func(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error) {
			return nil, errors.New("stop failed")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.RunStop(context.Background(), 1, "/commands/stop")
	if err == nil {
		t.Error("RunStop() should return error when ExecuteCommand fails")
	}
}

func TestClient_RunStop_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "invalid",
		Endpoint:  "localhost:6254",
	}

	err := c.RunStop(context.Background(), 1, "/commands/stop")
	if err == nil {
		t.Error("RunStop() should return error when ensureConn fails")
	}
}

func TestClient_GetStringValue_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "websocket", // Invalid transport
		Endpoint:  "localhost:6254",
	}

	_, err := c.GetStringValue(context.Background(), 1, "/test/param")
	if err == nil {
		t.Error("GetStringValue() should return error when ensureConn fails")
	}
}

func TestClient_GetStringValue_GetValueError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("parameter not found")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	_, err := c.GetStringValue(context.Background(), 1, "/test/param")
	if err == nil {
		t.Error("GetStringValue() should return error when GetValue fails")
	}
}

func TestClient_GetStringValue_EmptyOID(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			// Verify OID was normalized to "/"
			if in.Oid != "/" {
				t.Errorf("Expected OID to be normalized to /, got %s", in.Oid)
			}
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "test"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	_, err := c.GetStringValue(context.Background(), 1, "")
	if err != nil {
		t.Errorf("GetStringValue() with empty OID should work, got error: %v", err)
	}
}

func TestClient_GetStringValue_BoolValue(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			// Return a bool value (encoded as int32)
			return &st2138pb.Value{
				Kind: &st2138pb.Value_Int32Value{Int32Value: 1},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	got, err := c.GetStringValue(context.Background(), 1, "/test/bool")
	if err != nil {
		t.Errorf("GetStringValue() error = %v", err)
	}

	if got != "1" {
		t.Errorf("GetStringValue() with int32 value = %s, want '1'", got)
	}
}

func TestClient_GetStringValue_ZeroSlot(t *testing.T) {
	callCount := 0
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			callCount++
			if in.Slot != 0 {
				t.Errorf("Expected slot 0, got %d", in.Slot)
			}
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "value"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	_, err := c.GetStringValue(context.Background(), 0, "/test")
	if err != nil {
		t.Errorf("GetStringValue() with slot 0 error = %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 GetValue call, got %d", callCount)
	}
}
