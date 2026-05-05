package client

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
)

// Additional tests for params operations to improve coverage

func TestClient_SetStringValue_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "mqtt", // Invalid transport
		Endpoint:  "localhost:6254",
	}

	err := c.SetStringValue(context.Background(), 1, "/test/param", "value")
	if err == nil {
		t.Error("SetStringValue() should return error when ensureConn fails")
	}
}

func TestClient_SetStringValue_SetValueError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("write failed")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetStringValue(context.Background(), 1, "/test/param", "value")
	if err == nil {
		t.Error("SetStringValue() should return error when SetValue fails")
	}
}

func TestClient_SetStringValue_EmptyValue(t *testing.T) {
	var capturedValue string
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			capturedValue = in.Value.Value.GetStringValue()
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetStringValue(context.Background(), 1, "/test/param", "")
	if err != nil {
		t.Errorf("SetStringValue() with empty value error = %v", err)
	}

	if capturedValue != "" {
		t.Errorf("Expected empty string to be set, got %s", capturedValue)
	}
}

func TestClient_SetStringValue_LongValue(t *testing.T) {
	longValue := string(make([]byte, 10000))
	for i := range longValue {
		longValue = longValue[:i] + "x" + longValue[i+1:]
	}

	var capturedLength int
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			capturedLength = len(in.Value.Value.GetStringValue())
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetStringValue(context.Background(), 1, "/test/param", longValue)
	if err != nil {
		t.Errorf("SetStringValue() with long value error = %v", err)
	}

	if capturedLength != 10000 {
		t.Errorf("Expected long value to be preserved, got length %d", capturedLength)
	}
}

func TestClient_SetStringValue_SpecialCharacters(t *testing.T) {
	specialValue := "test\n\r\t\"'\\unicode: \u4e2d\u6587"

	var capturedValue string
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			capturedValue = in.Value.Value.GetStringValue()
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetStringValue(context.Background(), 1, "/test/param", specialValue)
	if err != nil {
		t.Errorf("SetStringValue() with special chars error = %v", err)
	}

	if capturedValue != specialValue {
		t.Errorf("Special characters not preserved, got %s", capturedValue)
	}
}

func TestClient_SetNumberValue_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "tcp", // Invalid transport
		Endpoint:  "localhost:6254",
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", 42.0)
	if err == nil {
		t.Error("SetNumberValue() should return error when ensureConn fails")
	}
}

func TestClient_SetNumberValue_SetValueError(t *testing.T) {
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("device offline")
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", 99.9)
	if err == nil {
		t.Error("SetNumberValue() should return error when SetValue fails")
	}
}

func TestClient_SetNumberValue_Zero(t *testing.T) {
	var wasInt32 bool
	var capturedValue int32
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if v, ok := in.Value.Value.Kind.(*st2138pb.Value_Int32Value); ok {
				wasInt32 = true
				capturedValue = v.Int32Value
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", 0.0)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasInt32 {
		t.Error("Zero should be sent as int32")
	}
	if capturedValue != 0 {
		t.Errorf("Expected 0, got %d", capturedValue)
	}
}

func TestClient_SetNumberValue_NegativeInteger(t *testing.T) {
	var wasInt32 bool
	var capturedValue int32
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if v, ok := in.Value.Value.Kind.(*st2138pb.Value_Int32Value); ok {
				wasInt32 = true
				capturedValue = v.Int32Value
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", -123.0)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasInt32 {
		t.Error("Negative integer should be sent as int32")
	}
	if capturedValue != -123 {
		t.Errorf("Expected -123, got %d", capturedValue)
	}
}

func TestClient_SetNumberValue_MaxInt32(t *testing.T) {
	var wasInt32 bool
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if _, ok := in.Value.Value.Kind.(*st2138pb.Value_Int32Value); ok {
				wasInt32 = true
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", math.MaxInt32)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasInt32 {
		t.Error("MaxInt32 should be sent as int32")
	}
}

func TestClient_SetNumberValue_MinInt32(t *testing.T) {
	var wasInt32 bool
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if _, ok := in.Value.Value.Kind.(*st2138pb.Value_Int32Value); ok {
				wasInt32 = true
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", math.MinInt32)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasInt32 {
		t.Error("MinInt32 should be sent as int32")
	}
}

func TestClient_SetNumberValue_JustOverMaxInt32(t *testing.T) {
	var wasFloat32 bool
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if _, ok := in.Value.Value.Kind.(*st2138pb.Value_Float32Value); ok {
				wasFloat32 = true
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", float64(math.MaxInt32)+1)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasFloat32 {
		t.Error("MaxInt32+1 should be sent as float32")
	}
}

func TestClient_SetNumberValue_JustUnderMinInt32(t *testing.T) {
	var wasFloat32 bool
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if _, ok := in.Value.Value.Kind.(*st2138pb.Value_Float32Value); ok {
				wasFloat32 = true
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", float64(math.MinInt32)-1)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasFloat32 {
		t.Error("MinInt32-1 should be sent as float32")
	}
}

func TestClient_SetNumberValue_VerySmallDecimal(t *testing.T) {
	var wasFloat32 bool
	var capturedValue float32
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if v, ok := in.Value.Value.Kind.(*st2138pb.Value_Float32Value); ok {
				wasFloat32 = true
				capturedValue = v.Float32Value
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", 0.0001)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasFloat32 {
		t.Error("Small decimal should be sent as float32")
	}
	if capturedValue == 0 {
		t.Error("Small decimal value lost precision")
	}
}

func TestClient_SetNumberValue_NegativeFloat(t *testing.T) {
	var wasFloat32 bool
	var capturedValue float32
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if v, ok := in.Value.Value.Kind.(*st2138pb.Value_Float32Value); ok {
				wasFloat32 = true
				capturedValue = v.Float32Value
			}
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "/test/number", -3.14159)
	if err != nil {
		t.Errorf("SetNumberValue() error = %v", err)
	}

	if !wasFloat32 {
		t.Error("Negative float should be sent as float32")
	}
	if capturedValue >= 0 {
		t.Errorf("Expected negative value, got %f", capturedValue)
	}
}

func TestClient_SetParams_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "rest", // Invalid
		Endpoint:  "localhost:6254",
	}

	err := c.SetParams(context.Background(), types.DynamicNull())
	if err == nil {
		t.Error("SetParams() should return error when ensureConn fails")
	}
}

func TestClient_SetParamsWithSlot_EnsureConnError(t *testing.T) {
	c := &Client{
		Transport: "jsonrpc", // Invalid
		Endpoint:  "localhost:6254",
	}

	err := c.SetParamsWithSlot(context.Background(), types.DynamicNull(), 1)
	if err == nil {
		t.Error("SetParamsWithSlot() should return error when ensureConn fails")
	}
}

func TestClient_SetNumberValue_EmptyOID(t *testing.T) {
	var capturedOID string
	mockClient := &mockCatenaServiceClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			capturedOID = in.Value.Oid
			return &st2138pb.Empty{}, nil
		},
	}

	c := &Client{
		Transport: "grpc",
		rpcClient: mockClient,
		conn:      &grpc.ClientConn{},
	}

	err := c.SetNumberValue(context.Background(), 1, "", 42.0)
	if err != nil {
		t.Errorf("SetNumberValue() with empty OID error = %v", err)
	}

	// Empty OID should be normalized to "/"
	if capturedOID != "/" {
		t.Errorf("Expected OID to be normalized to /, got %s", capturedOID)
	}
}
