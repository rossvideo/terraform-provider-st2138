package device

import (
	"crypto/sha256"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

func TestToFOID(t *testing.T) {
	tests := map[string]string{
		"authz_monitor":       "/authz_monitor",
		"/authz_monitor":      "/authz_monitor",
		"  counter ":          "/counter",
		"constraint/examples": "/constraint/examples",
		"":                    "",
	}
	for in, want := range tests {
		if got := toFOID(in); got != want {
			t.Errorf("toFOID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChangedParams(t *testing.T) {
	prev := map[string]attr.Value{
		"counter":        types.NumberValue(big.NewFloat(1)),
		"string_example": types.StringValue("Hello"),
		"removed":        types.StringValue("x"),
	}
	next := map[string]attr.Value{
		"counter":        types.NumberValue(big.NewFloat(1)),
		"string_example": types.StringValue("World"),
		"added":          types.StringValue("y"),
	}

	got := changedParams(prev, next)

	if len(got) != 2 {
		t.Fatalf("changed = %v, want string_example and added", got)
	}
	if _, ok := got["string_example"]; !ok {
		t.Error("string_example should be changed")
	}
	if _, ok := got["added"]; !ok {
		t.Error("added should be changed")
	}
}

func TestCommandInvocationFromAttr_NormalizesOIDs(t *testing.T) {
	r := &deviceResource{}

	inv, err := r.commandInvocationFromAttr(types.StringValue("fib_stop"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.OID != "/fib_stop" {
		t.Errorf("OID = %q, want /fib_stop", inv.OID)
	}

	obj := types.ObjectValueMust(
		map[string]attr.Type{"command": types.StringType, "status_foid": types.StringType},
		map[string]attr.Value{"command": types.StringValue("fib_start"), "status_foid": types.StringValue("number_example")},
	)
	inv, err = r.commandInvocationFromAttr(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.OID != "/fib_start" {
		t.Errorf("OID = %q, want /fib_start", inv.OID)
	}
	if inv.StatusFoid != "/number_example" {
		t.Errorf("StatusFoid = %q, want /number_example", inv.StatusFoid)
	}
}

func TestAttrValueToProtoValue_StructVariant(t *testing.T) {
	r := &deviceResource{}
	variant := types.ObjectValueMust(
		map[string]attr.Type{"struct_variant_type": types.StringType, "value": types.NumberType},
		map[string]attr.Value{
			"struct_variant_type": types.StringValue("int_kind"),
			"value":               types.NumberValue(big.NewFloat(42)),
		},
	)
	variantObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"struct_variant_type": types.StringType,
		"value":               types.NumberType,
	}}
	input := types.ObjectValueMust(
		map[string]attr.Type{"nested_struct": variantObjectType},
		map[string]attr.Value{"nested_struct": variant},
	)
	descriptor := &st2138pb.Param{
		Type: st2138pb.ParamType_STRUCT,
		Params: map[string]*st2138pb.Param{
			"nested_struct": {
				Type: st2138pb.ParamType_STRUCT_VARIANT,
				Params: map[string]*st2138pb.Param{
					"int_kind": {Type: st2138pb.ParamType_INT32},
				},
			},
		},
	}

	got, err := r.attrValueToProtoValue(input, descriptor)
	if err != nil {
		t.Fatalf("attrValueToProtoValue() error = %v", err)
	}
	structValue := got.GetStructValue()
	if structValue == nil {
		t.Fatalf("value kind = %T, want struct", got.GetKind())
	}
	gotVariant := structValue.GetFields()["nested_struct"].GetStructVariantValue()
	if gotVariant == nil {
		t.Fatalf("nested_struct kind = %T, want struct variant", structValue.GetFields()["nested_struct"].GetKind())
	}
	if gotVariant.GetStructVariantType() != "int_kind" {
		t.Errorf("struct_variant_type = %q, want int_kind", gotVariant.GetStructVariantType())
	}
	if gotVariant.GetValue().GetInt32Value() != 42 {
		t.Errorf("variant int32 value = %d, want 42", gotVariant.GetValue().GetInt32Value())
	}
}

func TestAttrValueToProtoValue_StructVariantNestedStructWrapper(t *testing.T) {
	r := &deviceResource{}
	variantObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"struct_variant_type": types.StringType,
		"value":               types.NumberType,
	}}
	variant := types.ObjectValueMust(variantObjectType.AttrTypes, map[string]attr.Value{
		"struct_variant_type": types.StringValue("int_kind"),
		"value":               types.NumberValue(big.NewFloat(42)),
	})
	input := types.ObjectValueMust(
		map[string]attr.Type{"nested_struct": variantObjectType},
		map[string]attr.Value{"nested_struct": variant},
	)
	descriptor := &st2138pb.Param{
		Type: st2138pb.ParamType_STRUCT_VARIANT,
		Params: map[string]*st2138pb.Param{
			"int_kind": {Type: st2138pb.ParamType_INT32},
		},
	}

	got, err := r.attrValueToProtoValue(input, descriptor)
	if err != nil {
		t.Fatalf("attrValueToProtoValue() error = %v", err)
	}
	gotVariant := got.GetStructVariantValue()
	if gotVariant == nil {
		t.Fatalf("value kind = %T, want struct variant", got.GetKind())
	}
	if gotVariant.GetStructVariantType() != "int_kind" {
		t.Errorf("struct_variant_type = %q, want int_kind", gotVariant.GetStructVariantType())
	}
	if gotVariant.GetValue().GetInt32Value() != 42 {
		t.Errorf("variant int32 value = %d, want 42", gotVariant.GetValue().GetInt32Value())
	}
}

func TestAttrValueToProtoValue_StructVariantArray(t *testing.T) {
	r := &deviceResource{}
	variantObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"struct_variant_type": types.StringType,
		"value":               types.DynamicType,
	}}
	intVariant := types.ObjectValueMust(variantObjectType.AttrTypes, map[string]attr.Value{
		"struct_variant_type": types.StringValue("int_kind"),
		"value":               types.DynamicValue(types.NumberValue(big.NewFloat(42))),
	})
	stringVariant := types.ObjectValueMust(variantObjectType.AttrTypes, map[string]attr.Value{
		"struct_variant_type": types.StringValue("string_kind"),
		"value":               types.DynamicValue(types.StringValue("hello")),
	})
	variants := types.TupleValueMust(
		[]attr.Type{variantObjectType, variantObjectType},
		[]attr.Value{intVariant, stringVariant},
	)
	input := types.ObjectValueMust(
		map[string]attr.Type{"nested_struct": types.TupleType{ElemTypes: []attr.Type{variantObjectType, variantObjectType}}},
		map[string]attr.Value{"nested_struct": variants},
	)
	descriptor := &st2138pb.Param{
		Type: st2138pb.ParamType_STRUCT,
		Params: map[string]*st2138pb.Param{
			"nested_struct": {
				Type: st2138pb.ParamType_STRUCT_VARIANT_ARRAY,
				Params: map[string]*st2138pb.Param{
					"int_kind":    {Type: st2138pb.ParamType_INT32},
					"string_kind": {Type: st2138pb.ParamType_STRING},
				},
			},
		},
	}

	got, err := r.attrValueToProtoValue(input, descriptor)
	if err != nil {
		t.Fatalf("attrValueToProtoValue() error = %v", err)
	}
	structValue := got.GetStructValue()
	if structValue == nil {
		t.Fatalf("value kind = %T, want struct", got.GetKind())
	}
	gotVariants := structValue.GetFields()["nested_struct"].GetStructVariantArrayValues().GetStructVariants()
	if len(gotVariants) != 2 {
		t.Fatalf("variant count = %d, want 2", len(gotVariants))
	}
	if gotVariants[0].GetValue().GetInt32Value() != 42 {
		t.Errorf("first variant int32 value = %d, want 42", gotVariants[0].GetValue().GetInt32Value())
	}
	if gotVariants[1].GetValue().GetStringValue() != "hello" {
		t.Errorf("second variant string value = %q, want hello", gotVariants[1].GetValue().GetStringValue())
	}
}

func TestAttrValueToProtoValue_StructVariantArrayNestedStructWrapper(t *testing.T) {
	r := &deviceResource{}
	variantObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"struct_variant_type": types.StringType,
		"value":               types.DynamicType,
	}}
	intVariant := types.ObjectValueMust(variantObjectType.AttrTypes, map[string]attr.Value{
		"struct_variant_type": types.StringValue("int_kind"),
		"value":               types.DynamicValue(types.NumberValue(big.NewFloat(42))),
	})
	stringVariant := types.ObjectValueMust(variantObjectType.AttrTypes, map[string]attr.Value{
		"struct_variant_type": types.StringValue("string_kind"),
		"value":               types.DynamicValue(types.StringValue("hello")),
	})
	wrapperType := types.ObjectType{AttrTypes: map[string]attr.Type{"nested_struct": variantObjectType}}
	intWrapper := types.ObjectValueMust(wrapperType.AttrTypes, map[string]attr.Value{"nested_struct": intVariant})
	stringWrapper := types.ObjectValueMust(wrapperType.AttrTypes, map[string]attr.Value{"nested_struct": stringVariant})
	input := types.TupleValueMust(
		[]attr.Type{wrapperType, wrapperType},
		[]attr.Value{intWrapper, stringWrapper},
	)
	descriptor := &st2138pb.Param{
		Type: st2138pb.ParamType_STRUCT_VARIANT_ARRAY,
		Params: map[string]*st2138pb.Param{
			"int_kind":    {Type: st2138pb.ParamType_INT32},
			"string_kind": {Type: st2138pb.ParamType_STRING},
		},
	}

	got, err := r.attrValueToProtoValue(input, descriptor)
	if err != nil {
		t.Fatalf("attrValueToProtoValue() error = %v", err)
	}
	gotVariants := got.GetStructVariantArrayValues().GetStructVariants()
	if len(gotVariants) != 2 {
		t.Fatalf("variant count = %d, want 2", len(gotVariants))
	}
	if gotVariants[0].GetValue().GetInt32Value() != 42 {
		t.Errorf("first variant int32 value = %d, want 42", gotVariants[0].GetValue().GetInt32Value())
	}
	if gotVariants[1].GetValue().GetStringValue() != "hello" {
		t.Errorf("second variant string value = %q, want hello", gotVariants[1].GetValue().GetStringValue())
	}
}

func TestAttrValueToProtoValue_BinaryBase64Payload(t *testing.T) {
	r := &deviceResource{}
	dataPayloadType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"metadata":         types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"digest":           types.StringType,
		"payload_encoding": types.StringType,
		"payload":          types.StringType,
	}}
	dataPayload := types.ObjectValueMust(dataPayloadType.AttrTypes, map[string]attr.Value{
		"metadata":         types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		"digest":           types.StringValue(""),
		"payload_encoding": types.StringValue("UNCOMPRESSED"),
		"payload":          types.StringValue("yv66vg=="),
	})
	input := types.ObjectValueMust(
		map[string]attr.Type{"data_payload": dataPayloadType},
		map[string]attr.Value{"data_payload": dataPayload},
	)

	got, err := r.attrValueToProtoValue(input, &st2138pb.Param{Type: st2138pb.ParamType_BINARY})
	if err != nil {
		t.Fatalf("attrValueToProtoValue() error = %v", err)
	}
	gotPayload := got.GetDataPayload()
	if gotPayload == nil {
		t.Fatalf("value kind = %T, want data payload", got.GetKind())
	}
	if string(gotPayload.GetPayload()) != string([]byte{0xca, 0xfe, 0xba, 0xbe}) {
		t.Errorf("payload = %v, want class-file magic bytes", gotPayload.GetPayload())
	}
	if gotPayload.GetPayloadEncoding() != st2138pb.DataPayload_UNCOMPRESSED {
		t.Errorf("payload encoding = %v, want UNCOMPRESSED", gotPayload.GetPayloadEncoding())
	}
	if len(gotPayload.GetDigest()) != sha256.Size {
		t.Errorf("digest length = %d, want %d", len(gotPayload.GetDigest()), sha256.Size)
	}
}

func TestAttrValueToProtoValue_BinaryFilePayload(t *testing.T) {
	r := &deviceResource{}
	filePath := filepath.Join(t.TempDir(), "payload.bin")
	contents := []byte{0, 1, 2, 255}
	if err := os.WriteFile(filePath, contents, 0o600); err != nil {
		t.Fatalf("write test payload: %v", err)
	}
	dataPayloadType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"payload_file": types.StringType,
	}}
	dataPayload := types.ObjectValueMust(dataPayloadType.AttrTypes, map[string]attr.Value{
		"payload_file": types.StringValue(filePath),
	})
	input := types.ObjectValueMust(
		map[string]attr.Type{"data_payload": dataPayloadType},
		map[string]attr.Value{"data_payload": dataPayload},
	)

	got, err := r.attrValueToProtoValue(input, &st2138pb.Param{Type: st2138pb.ParamType_BINARY})
	if err != nil {
		t.Fatalf("attrValueToProtoValue() error = %v", err)
	}
	gotPayload := got.GetDataPayload()
	if gotPayload == nil {
		t.Fatalf("value kind = %T, want data payload", got.GetKind())
	}
	if string(gotPayload.GetPayload()) != string(contents) {
		t.Errorf("payload = %v, want %v", gotPayload.GetPayload(), contents)
	}
	wantDigest := sha256.Sum256(contents)
	if string(gotPayload.GetDigest()) != string(wantDigest[:]) {
		t.Errorf("digest = %x, want %x", gotPayload.GetDigest(), wantDigest)
	}
}

func TestAttrValueToProtoValue_BinaryPayloadRequiresSingleSource(t *testing.T) {
	r := &deviceResource{}
	input := types.ObjectValueMust(map[string]attr.Type{
		"payload":      types.StringType,
		"payload_file": types.StringType,
	}, map[string]attr.Value{
		"payload":      types.StringValue("YQ=="),
		"payload_file": types.StringValue("payload.bin"),
	})

	_, err := r.attrValueToProtoValue(input, &st2138pb.Param{Type: st2138pb.ParamType_BINARY})
	if err == nil {
		t.Fatal("attrValueToProtoValue() error = nil, want error for multiple payload sources")
	}
	if err.Error() != "data payload must set exactly one of payload, payload_file, or url" {
		t.Errorf("error = %q, want multiple payload source error", err.Error())
	}
}
