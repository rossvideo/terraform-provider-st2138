package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles struct array param creation in Terraform
// Read: Fetches current struct array param state from device
// Update: Modifies existing struct array params on device
// Delete: Removes struct array params from device

// SetStructArrayValue sets a struct array parameter value via gRPC.
// Struct arrays contain multiple structured parameter sets.
// The OID is normalized to ensure it starts with '/'.
func SetStructArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, structList *st2138pb.StructList) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_StructArrayValues{StructArrayValues: structList}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetStructArrayValue fetches a struct array parameter value via gRPC.
// Returns the struct array or an error if the fetch fails.
func GetStructArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.StructList, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetStructArrayValues(), nil
}
