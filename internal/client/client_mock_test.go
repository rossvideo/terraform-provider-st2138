package client

import (
	"context"
	"testing"
	"time"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
)

// Mock gRPC client for testing
type mockCatenaServiceClient struct {
	st2138pb.CatenaServiceClient
	setValueFunc       func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error)
	getValueFunc       func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error)
	executeCommandFunc func(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error)
}

func (m *mockCatenaServiceClient) SetValue(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
	if m.setValueFunc != nil {
		return m.setValueFunc(ctx, in, opts...)
	}
	return &st2138pb.Empty{}, nil
}

func (m *mockCatenaServiceClient) GetValue(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
	if m.getValueFunc != nil {
		return m.getValueFunc(ctx, in, opts...)
	}
	return &st2138pb.Value{}, nil
}

func (m *mockCatenaServiceClient) ExecuteCommand(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error) {
	if m.executeCommandFunc != nil {
		return m.executeCommandFunc(ctx, in, opts...)
	}
	return nil, nil
}

func TestClient_GetStringValue(t *testing.T) {
	tests := []struct {
		name      string
		slot      uint32
		oid       string
		mockValue *st2138pb.Value
		want      string
		wantErr   bool
	}{
		{
			name: "string value",
			slot: 1,
			oid:  "/test/param",
			mockValue: &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "test-value"},
			},
			want:    "test-value",
			wantErr: false,
		},
		{
			name: "int32 value",
			slot: 1,
			oid:  "/test/number",
			mockValue: &st2138pb.Value{
				Kind: &st2138pb.Value_Int32Value{Int32Value: 42},
			},
			want:    "42",
			wantErr: false,
		},
		{
			name: "float32 value",
			slot: 1,
			oid:  "/test/float",
			mockValue: &st2138pb.Value{
				Kind: &st2138pb.Value_Float32Value{Float32Value: 3.14},
			},
			want:    "3.14",
			wantErr: false,
		},
		{
			name: "OID without leading slash",
			slot: 1,
			oid:  "test/param",
			mockValue: &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "value"},
			},
			want:    "value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCatenaServiceClient{
				getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
					return tt.mockValue, nil
				},
			}

			c := &Client{
				Transport: "grpc",
				rpcClient: mockClient,
				conn:      &grpc.ClientConn{}, // Non-nil to skip dial
			}

			got, err := c.GetStringValue(context.Background(), tt.slot, tt.oid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStringValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStringValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SetStringValue(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			// Verify the payload structure
			if in.Slot != 1 {
				t.Errorf("Expected slot 1, got %d", in.Slot)
			}
			if in.Value.Oid != "/test/param" {
				t.Errorf("Expected OID /test/param, got %s", in.Value.Oid)
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetStringValue(context.Background(), 1, "/test/param", "test-value")
	if err != nil {
		t.Errorf("SetStringValue() error = %v", err)
	}
}

func TestClient_SetNumberValue(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		wantInt32 bool
	}{
		{
			name:      "integer value",
			value:     42.0,
			wantInt32: true,
		},
		{
			name:      "float value",
			value:     3.14,
			wantInt32: false,
		},
		{
			name:      "large integer",
			value:     2147483647.0,
			wantInt32: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCatenaServiceClient{
				setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
					if tt.wantInt32 {
						if _, ok := in.Value.Value.Kind.(*st2138pb.Value_Int32Value); !ok {
							t.Error("Expected Int32Value kind")
						}
					} else {
						if _, ok := in.Value.Value.Kind.(*st2138pb.Value_Float32Value); !ok {
							t.Error("Expected Float32Value kind")
						}
					}
					return &st2138pb.Empty{}, nil
				},
			}

			c := &Client{
				Transport: "grpc",
				rpcClient: mockClient,
				conn:      &grpc.ClientConn{},
			}

			err := c.SetNumberValue(context.Background(), 1, "/test/number", tt.value)
			if err != nil {
				t.Errorf("SetNumberValue() error = %v", err)
			}
		})
	}
}

func TestClient_WaitReady_Success(t *testing.T) {
	callCount := 0
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			callCount++
			// Return ready after 2 calls
			if callCount >= 2 {
				return &st2138pb.Value{
					Kind: &st2138pb.Value_StringValue{StringValue: "ready"},
				}, nil
			}
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "not-ready"},
			}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.WaitReady(context.Background(), 1, "/status", "ready", 5*time.Second)
	if err != nil {
		t.Errorf("WaitReady() error = %v", err)
	}
	if callCount < 2 {
		t.Errorf("Expected at least 2 calls, got %d", callCount)
	}
}

func TestClient_WaitNotReady_Success(t *testing.T) {
	callCount := 0
	mockClient := &mockCatenaServiceClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			callCount++
			// Return not-ready after 2 calls
			if callCount >= 2 {
				return &st2138pb.Value{
					Kind: &st2138pb.Value_StringValue{StringValue: "stopped"},
				}, nil
			}
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

	err := c.WaitNotReady(context.Background(), 1, "/status", "ready", 5*time.Second)
	if err != nil {
		t.Errorf("WaitNotReady() error = %v", err)
	}
	if callCount < 2 {
		t.Errorf("Expected at least 2 calls, got %d", callCount)
	}
}

func TestClient_RunStart(t *testing.T) {
	executeCalled := false
	mockClient := &mockCatenaServiceClient{
		executeCommandFunc: func(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error) {
			executeCalled = true
			// Verify the payload
			if in.Slot != 1 {
				t.Errorf("Expected slot 1, got %d", in.Slot)
			}
			if in.Oid != "/commands/start" {
				t.Errorf("Expected OID /commands/start, got %s", in.Oid)
			}
			if in.Respond != false {
				t.Error("Expected Respond to be false")
			}
			return nil, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.RunStart(context.Background(), 1, "/commands/start")
	if err != nil {
		t.Errorf("RunStart() error = %v", err)
	}
	if !executeCalled {
		t.Error("ExecuteCommand should have been called")
	}
}

func TestClient_RunStart_WithoutLeadingSlash(t *testing.T) {
	executeCalled := false
	mockClient := &mockCatenaServiceClient{
		executeCommandFunc: func(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error) {
			executeCalled = true
			// Should normalize OID
			if in.Oid != "/commands/start" {
				t.Errorf("Expected normalized OID /commands/start, got %s", in.Oid)
			}
			return nil, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.RunStart(context.Background(), 1, "commands/start")
	if err != nil {
		t.Errorf("RunStart() error = %v", err)
	}
	if !executeCalled {
		t.Error("ExecuteCommand should have been called")
	}
}

func TestClient_RunStop(t *testing.T) {
	executeCalled := false
	mockClient := &mockCatenaServiceClient{
		executeCommandFunc: func(ctx context.Context, in *st2138pb.ExecuteCommandPayload, opts ...grpc.CallOption) (st2138pb.CatenaService_ExecuteCommandClient, error) {
			executeCalled = true
			if in.Slot != 1 {
				t.Errorf("Expected slot 1, got %d", in.Slot)
			}
			if in.Oid != "/commands/stop" {
				t.Errorf("Expected OID /commands/stop, got %s", in.Oid)
			}
			return nil, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.RunStop(context.Background(), 1, "/commands/stop")
	if err != nil {
		t.Errorf("RunStop() error = %v", err)
	}
	if !executeCalled {
		t.Error("ExecuteCommand should have been called")
	}
}

func TestClient_SetStringValue_WithoutLeadingSlash(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			// Should normalize OID
			if in.Value.Oid != "/test/param" {
				t.Errorf("Expected normalized OID /test/param, got %s", in.Value.Oid)
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetStringValue(context.Background(), 1, "test/param", "value")
	if err != nil {
		t.Errorf("SetStringValue() error = %v", err)
	}
}

func TestClient_SetNumberValue_WithoutLeadingSlash(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if in.Value.Oid != "/test/number" {
				t.Errorf("Expected normalized OID /test/number, got %s", in.Value.Oid)
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "test/number", 42)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}
}
