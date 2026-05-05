package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles struct variant array param creation in Terraform
// Read: Fetches current struct variant array param state from device
// Update: Modifies existing struct variant array params on device
// Delete: Removes struct variant array params from device

// SetStructVariantArrayValue sets a struct variant array parameter value via gRPC.
// Struct variant arrays contain multiple polymorphic structured parameter sets.
// The OID is normalized to ensure it starts with '/'.
func SetStructVariantArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, variantList *st2138pb.StructVariantList) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_StructVariantArrayValues{StructVariantArrayValues: variantList}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetStructVariantArrayValue fetches a struct variant array parameter value via gRPC.
// Returns the struct variant array or an error if the fetch fails.
func GetStructVariantArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.StructVariantList, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetStructVariantArrayValues(), nil
}
