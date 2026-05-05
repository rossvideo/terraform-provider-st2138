package params

import (
	"context"
	"testing"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
)

// Mock gRPC client for params testing
type mockParamsClient struct {
	st2138pb.CatenaServiceClient
	setValueFunc func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error)
	getValueFunc func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error)
}

func (m *mockParamsClient) SetValue(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
	if m.setValueFunc != nil {
		return m.setValueFunc(ctx, in, opts...)
	}
	return &st2138pb.Empty{}, nil
}

func (m *mockParamsClient) GetValue(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
	if m.getValueFunc != nil {
		return m.getValueFunc(ctx, in, opts...)
	}
	return &st2138pb.Value{}, nil
}

func TestSetStringValue(t *testing.T) {
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if in.Slot != 1 {
				t.Errorf("Expected slot 1, got %d", in.Slot)
			}
			if in.Value.Oid != "/test/param" {
				t.Errorf("Expected OID /test/param, got %s", in.Value.Oid)
			}
			if sv := in.Value.Value.GetStringValue(); sv != "test-value" {
				t.Errorf("Expected value 'test-value', got %s", sv)
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStringValue(context.Background(), mockClient, 1, "/test/param", "test-value")
	if err != nil {
		t.Errorf("SetStringValue() error = %v", err)
	}
}

func TestSetStringValue_OIDNormalization(t *testing.T) {
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if in.Value.Oid != "/test/param" {
				t.Errorf("Expected normalized OID /test/param, got %s", in.Value.Oid)
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStringValue(context.Background(), mockClient, 1, "test/param", "value")
	if err != nil {
		t.Errorf("SetStringValue() error = %v", err)
	}
}

func TestGetStringValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringValue{StringValue: "retrieved-value"},
			}, nil
		},
	}

	got, err := GetStringValue(context.Background(), mockClient, 1, "/test/param")
	if err != nil {
		t.Errorf("GetStringValue() error = %v", err)
	}
	if got != "retrieved-value" {
		t.Errorf("GetStringValue() = %s, want retrieved-value", got)
	}
}

func TestSetInt32Value(t *testing.T) {
	tests := []struct {
		name  string
		value int32
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -42},
		{"max", 2147483647},
		{"min", -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockParamsClient{
				setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
					if iv := in.Value.Value.GetInt32Value(); iv != tt.value {
						t.Errorf("Expected value %d, got %d", tt.value, iv)
					}
					return &st2138pb.Empty{}, nil
				},
			}

			err := SetInt32Value(context.Background(), mockClient, 1, "/test/int", tt.value)
			if err != nil {
				t.Errorf("SetInt32Value() error = %v", err)
			}
		})
	}
}

func TestGetInt32Value(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_Int32Value{Int32Value: 123},
			}, nil
		},
	}

	got, err := GetInt32Value(context.Background(), mockClient, 1, "/test/int")
	if err != nil {
		t.Errorf("GetInt32Value() error = %v", err)
	}
	if got != 123 {
		t.Errorf("GetInt32Value() = %d, want 123", got)
	}
}

func TestSetFloat32Value(t *testing.T) {
	tests := []struct {
		name  string
		value float32
	}{
		{"zero", 0.0},
		{"positive", 3.14},
		{"negative", -2.5},
		{"small", 0.0001},
		{"large", 999999.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockParamsClient{
				setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
					if fv := in.Value.Value.GetFloat32Value(); fv != tt.value {
						t.Errorf("Expected value %f, got %f", tt.value, fv)
					}
					return &st2138pb.Empty{}, nil
				},
			}

			err := SetFloat32Value(context.Background(), mockClient, 1, "/test/float", tt.value)
			if err != nil {
				t.Errorf("SetFloat32Value() error = %v", err)
			}
		})
	}
}

func TestGetFloat32Value(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_Float32Value{Float32Value: 3.14},
			}, nil
		},
	}

	got, err := GetFloat32Value(context.Background(), mockClient, 1, "/test/float")
	if err != nil {
		t.Errorf("GetFloat32Value() error = %v", err)
	}
	if got != 3.14 {
		t.Errorf("GetFloat32Value() = %f, want 3.14", got)
	}
}

func TestSetEmptyValue(t *testing.T) {
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if _, ok := in.Value.Value.Kind.(*st2138pb.Value_EmptyValue); !ok {
				t.Error("Expected EmptyValue kind")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetEmptyValue(context.Background(), mockClient, 1, "/test/trigger")
	if err != nil {
		t.Errorf("SetEmptyValue() error = %v", err)
	}
}

func TestSetInt32ArrayValue(t *testing.T) {
	values := &st2138pb.Int32List{Ints: []int32{1, 2, 3, 4, 5}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			arr := in.Value.Value.GetInt32ArrayValues()
			if arr == nil {
				t.Error("Expected Int32ArrayValues")
				return &st2138pb.Empty{}, nil
			}
			if len(arr.Ints) != 5 {
				t.Errorf("Expected 5 values, got %d", len(arr.Ints))
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetInt32ArrayValue(context.Background(), mockClient, 1, "/test/array", values)
	if err != nil {
		t.Errorf("SetInt32ArrayValue() error = %v", err)
	}
}

func TestGetInt32ArrayValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_Int32ArrayValues{
					Int32ArrayValues: &st2138pb.Int32List{Ints: []int32{1, 2, 3}},
				},
			}, nil
		},
	}

	got, err := GetInt32ArrayValue(context.Background(), mockClient, 1, "/test/array")
	if err != nil {
		t.Errorf("GetInt32ArrayValue() error = %v", err)
	}
	if got == nil || len(got.Ints) != 3 {
		t.Errorf("Expected array with 3 values, got %v", got)
	}
}

func TestSetFloat32ArrayValue(t *testing.T) {
	values := &st2138pb.Float32List{Floats: []float32{1.1, 2.2, 3.3}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			arr := in.Value.Value.GetFloat32ArrayValues()
			if arr == nil || len(arr.Floats) != 3 {
				t.Error("Expected Float32ArrayValues with 3 elements")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetFloat32ArrayValue(context.Background(), mockClient, 1, "/test/floats", values)
	if err != nil {
		t.Errorf("SetFloat32ArrayValue() error = %v", err)
	}
}

func TestSetStringArrayValue(t *testing.T) {
	values := &st2138pb.StringList{Strings: []string{"a", "b", "c"}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			arr := in.Value.Value.GetStringArrayValues()
			if arr == nil || len(arr.Strings) != 3 {
				t.Error("Expected StringArrayValues with 3 elements")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStringArrayValue(context.Background(), mockClient, 1, "/test/strings", values)
	if err != nil {
		t.Errorf("SetStringArrayValue() error = %v", err)
	}
}

func TestSetStructValue(t *testing.T) {
	structVal := &st2138pb.StructValue{}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if in.Value.Value.GetStructValue() == nil {
				t.Error("Expected StructValue")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStructValue(context.Background(), mockClient, 1, "/test/struct", structVal)
	if err != nil {
		t.Errorf("SetStructValue() error = %v", err)
	}
}

func TestSetDataValue(t *testing.T) {
	dataPayload := &st2138pb.DataPayload{
		Metadata: map[string]string{"type": "test"},
	}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			if in.Value.Value.GetDataPayload() == nil {
				t.Error("Expected DataPayload")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetDataValue(context.Background(), mockClient, 1, "/test/data", dataPayload)
	if err != nil {
		t.Errorf("SetDataValue() error = %v", err)
	}
}
