package params

import (
	"context"
	"errors"
	"testing"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
)

// Tests to achieve 100% coverage on all Get functions

func TestGetFloat32ArrayValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_Float32ArrayValues{
					Float32ArrayValues: &st2138pb.Float32List{Floats: []float32{1.1, 2.2, 3.3}},
				},
			}, nil
		},
	}

	got, err := GetFloat32ArrayValue(context.Background(), mockClient, 1, "/test/array")
	if err != nil {
		t.Errorf("GetFloat32ArrayValue() error = %v", err)
	}
	if got == nil || len(got.Floats) != 3 {
		t.Errorf("Expected array with 3 values, got %v", got)
	}
}

func TestGetFloat32ArrayValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("get failed")
		},
	}

	_, err := GetFloat32ArrayValue(context.Background(), mockClient, 1, "/test/array")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestGetStringArrayValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StringArrayValues{
					StringArrayValues: &st2138pb.StringList{Strings: []string{"a", "b", "c"}},
				},
			}, nil
		},
	}

	got, err := GetStringArrayValue(context.Background(), mockClient, 1, "/test/strings")
	if err != nil {
		t.Errorf("GetStringArrayValue() error = %v", err)
	}
	if got == nil || len(got.Strings) != 3 {
		t.Errorf("Expected array with 3 values, got %v", got)
	}
}

func TestGetStringArrayValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("connection lost")
		},
	}

	_, err := GetStringArrayValue(context.Background(), mockClient, 1, "/test/strings")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestGetDataValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_DataPayload{
					DataPayload: &st2138pb.DataPayload{
						Metadata: map[string]string{"type": "test"},
					},
				},
			}, nil
		},
	}

	got, err := GetDataValue(context.Background(), mockClient, 1, "/test/data")
	if err != nil {
		t.Errorf("GetDataValue() error = %v", err)
	}
	if got == nil {
		t.Errorf("Expected data payload, got nil")
	}
}

func TestGetDataValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("data unavailable")
		},
	}

	_, err := GetDataValue(context.Background(), mockClient, 1, "/test/data")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestGetStructValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StructValue{
					StructValue: &st2138pb.StructValue{},
				},
			}, nil
		},
	}

	got, err := GetStructValue(context.Background(), mockClient, 1, "/test/struct")
	if err != nil {
		t.Errorf("GetStructValue() error = %v", err)
	}
	if got == nil {
		t.Error("Expected struct value, got nil")
	}
}

func TestGetStructValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("struct not found")
		},
	}

	_, err := GetStructValue(context.Background(), mockClient, 1, "/test/struct")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestSetStructArrayValue(t *testing.T) {
	structs := &st2138pb.StructList{
		StructValues: []*st2138pb.StructValue{
			{},
			{},
		},
	}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			arr := in.Value.Value.GetStructArrayValues()
			if arr == nil || len(arr.StructValues) != 2 {
				t.Error("Expected StructArrayValues with 2 elements")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStructArrayValue(context.Background(), mockClient, 1, "/test/structs", structs)
	if err != nil {
		t.Errorf("SetStructArrayValue() error = %v", err)
	}
}

func TestSetStructArrayValue_Error(t *testing.T) {
	structs := &st2138pb.StructList{StructValues: []*st2138pb.StructValue{{}}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("write failed")
		},
	}

	err := SetStructArrayValue(context.Background(), mockClient, 1, "/test/structs", structs)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestGetStructArrayValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StructArrayValues{
					StructArrayValues: &st2138pb.StructList{
						StructValues: []*st2138pb.StructValue{{}, {}},
					},
				},
			}, nil
		},
	}

	got, err := GetStructArrayValue(context.Background(), mockClient, 1, "/test/structs")
	if err != nil {
		t.Errorf("GetStructArrayValue() error = %v", err)
	}
	if got == nil || len(got.StructValues) != 2 {
		t.Errorf("Expected array with 2 structs, got %v", got)
	}
}

func TestGetStructArrayValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("array not found")
		},
	}

	_, err := GetStructArrayValue(context.Background(), mockClient, 1, "/test/structs")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestSetStructVariantValue(t *testing.T) {
	variant := &st2138pb.StructVariantValue{StructVariantType: "test_type"}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			sv := in.Value.Value.GetStructVariantValue()
			if sv == nil {
				t.Error("Expected StructVariantValue")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStructVariantValue(context.Background(), mockClient, 1, "/test/variant", variant)
	if err != nil {
		t.Errorf("SetStructVariantValue() error = %v", err)
	}
}

func TestSetStructVariantValue_Error(t *testing.T) {
	variant := &st2138pb.StructVariantValue{StructVariantType: "test_type"}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("variant write failed")
		},
	}

	err := SetStructVariantValue(context.Background(), mockClient, 1, "/test/variant", variant)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestGetStructVariantValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StructVariantValue{
					StructVariantValue: &st2138pb.StructVariantValue{StructVariantType: "test_type"},
				},
			}, nil
		},
	}

	got, err := GetStructVariantValue(context.Background(), mockClient, 1, "/test/variant")
	if err != nil {
		t.Errorf("GetStructVariantValue() error = %v", err)
	}
	if got == nil {
		t.Error("Expected variant value, got nil")
	}
}

func TestGetStructVariantValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("variant not found")
		},
	}

	_, err := GetStructVariantValue(context.Background(), mockClient, 1, "/test/variant")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestSetStructVariantArrayValue(t *testing.T) {
	variants := &st2138pb.StructVariantList{
		StructVariants: []*st2138pb.StructVariantValue{
			{StructVariantType: "type1"},
			{StructVariantType: "type2"},
		},
	}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			arr := in.Value.Value.GetStructVariantArrayValues()
			if arr == nil || len(arr.StructVariants) != 2 {
				t.Error("Expected StructVariantArrayValues with 2 elements")
			}
			return &st2138pb.Empty{}, nil
		},
	}

	err := SetStructVariantArrayValue(context.Background(), mockClient, 1, "/test/variants", variants)
	if err != nil {
		t.Errorf("SetStructVariantArrayValue() error = %v", err)
	}
}

func TestSetStructVariantArrayValue_Error(t *testing.T) {
	variants := &st2138pb.StructVariantList{StructVariants: []*st2138pb.StructVariantValue{{StructVariantType: "type1"}}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("variant array write failed")
		},
	}

	err := SetStructVariantArrayValue(context.Background(), mockClient, 1, "/test/variants", variants)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestGetStructVariantArrayValue(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return &st2138pb.Value{
				Kind: &st2138pb.Value_StructVariantArrayValues{
					StructVariantArrayValues: &st2138pb.StructVariantList{
						StructVariants: []*st2138pb.StructVariantValue{
							{StructVariantType: "type1"},
							{StructVariantType: "type2"},
						},
					},
				},
			}, nil
		},
	}

	got, err := GetStructVariantArrayValue(context.Background(), mockClient, 1, "/test/variants")
	if err != nil {
		t.Errorf("GetStructVariantArrayValue() error = %v", err)
	}
	if got == nil || len(got.StructVariants) != 2 {
		t.Errorf("Expected array with 2 variants, got %v", got)
	}
}

func TestGetStructVariantArrayValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("variant array not found")
		},
	}

	_, err := GetStructVariantArrayValue(context.Background(), mockClient, 1, "/test/variants")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

// Tests for error paths to reach 100% on existing functions

func TestSetDataValue_Error(t *testing.T) {
	dataPayload := &st2138pb.DataPayload{Metadata: map[string]string{"key": "value"}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("data write failed")
		},
	}

	err := SetDataValue(context.Background(), mockClient, 1, "/test/data", dataPayload)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestSetEmptyValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("empty write failed")
		},
	}

	err := SetEmptyValue(context.Background(), mockClient, 1, "/test/trigger")
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestSetFloat32Value_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("float write failed")
		},
	}

	err := SetFloat32Value(context.Background(), mockClient, 1, "/test/float", 3.14)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestGetFloat32Value_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("float read failed")
		},
	}

	_, err := GetFloat32Value(context.Background(), mockClient, 1, "/test/float")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestSetFloat32ArrayValue_Error(t *testing.T) {
	values := &st2138pb.Float32List{Floats: []float32{1.1}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("float array write failed")
		},
	}

	err := SetFloat32ArrayValue(context.Background(), mockClient, 1, "/test/floats", values)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestSetInt32Value_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("int write failed")
		},
	}

	err := SetInt32Value(context.Background(), mockClient, 1, "/test/int", 42)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestGetInt32Value_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("int read failed")
		},
	}

	_, err := GetInt32Value(context.Background(), mockClient, 1, "/test/int")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestSetInt32ArrayValue_Error(t *testing.T) {
	values := &st2138pb.Int32List{Ints: []int32{1}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("int array write failed")
		},
	}

	err := SetInt32ArrayValue(context.Background(), mockClient, 1, "/test/ints", values)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestGetInt32ArrayValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("int array read failed")
		},
	}

	_, err := GetInt32ArrayValue(context.Background(), mockClient, 1, "/test/ints")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestGetStringValue_Error(t *testing.T) {
	mockClient := &mockParamsClient{
		getValueFunc: func(ctx context.Context, in *st2138pb.GetValuePayload, opts ...grpc.CallOption) (*st2138pb.Value, error) {
			return nil, errors.New("string read failed")
		},
	}

	_, err := GetStringValue(context.Background(), mockClient, 1, "/test/string")
	if err == nil {
		t.Error("Expected error when GetValue fails")
	}
}

func TestSetStringArrayValue_Error(t *testing.T) {
	values := &st2138pb.StringList{Strings: []string{"a"}}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("string array write failed")
		},
	}

	err := SetStringArrayValue(context.Background(), mockClient, 1, "/test/strings", values)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}

func TestSetStructValue_Error(t *testing.T) {
	structVal := &st2138pb.StructValue{}
	mockClient := &mockParamsClient{
		setValueFunc: func(ctx context.Context, in *st2138pb.SingleSetValuePayload, opts ...grpc.CallOption) (*st2138pb.Empty, error) {
			return nil, errors.New("struct write failed")
		},
	}

	err := SetStructValue(context.Background(), mockClient, 1, "/test/struct", structVal)
	if err == nil {
		t.Error("Expected error when SetValue fails")
	}
}
