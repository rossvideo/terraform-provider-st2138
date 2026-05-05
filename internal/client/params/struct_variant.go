package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles struct variant param creation in Terraform
// Read: Fetches current struct variant param state from device
// Update: Modifies existing struct variant params on device
// Delete: Removes struct variant params from device

// SetStructVariantValue sets a struct variant parameter value via gRPC.
// Struct variants allow for polymorphic structured data with type discrimination.
// The OID is normalized to ensure it starts with '/'.
func SetStructVariantValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, variantData *st2138pb.StructVariantValue) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_StructVariantValue{StructVariantValue: variantData}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetStructVariantValue fetches a struct variant parameter value via gRPC.
// Returns the struct variant data or an error if the fetch fails.
func GetStructVariantValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.StructVariantValue, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetStructVariantValue(), nil
}
