package client

import (
	"reflect"
	"testing"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

func TestStringifyValueAndValueToAny(t *testing.T) {
	structValue := &st2138pb.StructValue{Fields: map[string]*st2138pb.Value{
		"count": {Kind: &st2138pb.Value_Int32Value{Int32Value: 4}},
		"label": {Kind: &st2138pb.Value_StringValue{StringValue: "entry"}},
	}}
	dataPayload := &st2138pb.DataPayload{
		Metadata:        map[string]string{"content-type": "application/octet-stream"},
		PayloadEncoding: st2138pb.DataPayload_UNCOMPRESSED,
		Kind:            &st2138pb.DataPayload_Url{Url: "https://example.test/payload"},
	}
	variant := &st2138pb.StructVariantValue{
		StructVariantType: "int_kind",
		Value:             &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: 42}},
	}

	tests := []struct {
		name       string
		value      *st2138pb.Value
		wantString string
		wantAny    any
	}{
		{name: "nil", wantString: "", wantAny: ""},
		{name: "undefined", value: &st2138pb.Value{}, wantString: "", wantAny: ""},
		{name: "string", value: &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: "hello"}}, wantString: "hello", wantAny: "hello"},
		{name: "int32", value: &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: 42}}, wantString: "42", wantAny: int32(42)},
		{name: "float32", value: &st2138pb.Value{Kind: &st2138pb.Value_Float32Value{Float32Value: 1.25}}, wantString: "1.25", wantAny: float32(1.25)},
		{name: "empty", value: &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}, wantString: "", wantAny: ""},
		{name: "string array", value: &st2138pb.Value{Kind: &st2138pb.Value_StringArrayValues{StringArrayValues: &st2138pb.StringList{Strings: []string{"a", "b"}}}}, wantString: `["a","b"]`, wantAny: []string{"a", "b"}},
		{name: "int32 array", value: &st2138pb.Value{Kind: &st2138pb.Value_Int32ArrayValues{Int32ArrayValues: &st2138pb.Int32List{Ints: []int32{2, 3}}}}, wantString: `[2,3]`, wantAny: []int32{2, 3}},
		{name: "float32 array", value: &st2138pb.Value{Kind: &st2138pb.Value_Float32ArrayValues{Float32ArrayValues: &st2138pb.Float32List{Floats: []float32{1.5, 2}}}}, wantString: `[1.5,2]`, wantAny: []float32{1.5, 2}},
		{name: "struct", value: &st2138pb.Value{Kind: &st2138pb.Value_StructValue{StructValue: structValue}}, wantString: `{"count":4,"label":"entry"}`, wantAny: map[string]any{"count": int32(4), "label": "entry"}},
		{name: "struct array", value: &st2138pb.Value{Kind: &st2138pb.Value_StructArrayValues{StructArrayValues: &st2138pb.StructList{StructValues: []*st2138pb.StructValue{structValue}}}}, wantString: `[{"count":4,"label":"entry"}]`, wantAny: []map[string]any{{"count": int32(4), "label": "entry"}}},
		{name: "struct variant", value: &st2138pb.Value{Kind: &st2138pb.Value_StructVariantValue{StructVariantValue: variant}}, wantString: `{"struct_variant_type":"int_kind","value":42}`, wantAny: map[string]any{"struct_variant_type": "int_kind", "value": int32(42)}},
		{name: "struct variant array", value: &st2138pb.Value{Kind: &st2138pb.Value_StructVariantArrayValues{StructVariantArrayValues: &st2138pb.StructVariantList{StructVariants: []*st2138pb.StructVariantValue{variant}}}}, wantString: `[{"struct_variant_type":"int_kind","value":42}]`, wantAny: []map[string]any{{"struct_variant_type": "int_kind", "value": int32(42)}}},
		{name: "data payload", value: &st2138pb.Value{Kind: &st2138pb.Value_DataPayload{DataPayload: dataPayload}}, wantString: `{"metadata":{"content-type":"application/octet-stream"},"payload_encoding":"UNCOMPRESSED","url":"https://example.test/payload"}`, wantAny: map[string]any{"metadata": map[string]string{"content-type": "application/octet-stream"}, "payload_encoding": "UNCOMPRESSED", "url": "https://example.test/payload"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := stringifyValue(test.value); got != test.wantString {
				t.Errorf("stringifyValue() = %q, want %q", got, test.wantString)
			}
			if got := valueToAny(test.value); !reflect.DeepEqual(got, test.wantAny) {
				t.Errorf("valueToAny() = %#v, want %#v", got, test.wantAny)
			}
		})
	}
}

func TestStructValueConversionNil(t *testing.T) {
	if got := stringifyStructValue(nil); got != "{}" {
		t.Errorf("stringifyStructValue(nil) = %q, want {}", got)
	}
	if got := structValueToMap(nil); !reflect.DeepEqual(got, map[string]any{}) {
		t.Errorf("structValueToMap(nil) = %#v, want empty map", got)
	}
}
