package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles struct param creation in Terraform
// Read: Fetches current struct param state from device
// Update: Modifies existing struct params on device
// Delete: Removes struct params from device

// SetStructValue sets a structured parameter value via gRPC.
// Struct parameters contain nested key-value pairs organized hierarchically.
// The OID is normalized to ensure it starts with '/'.
func SetStructValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, structData *st2138pb.StructValue) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_StructValue{StructValue: structData}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetStructValue fetches a structured parameter value via gRPC.
// Returns the struct data or an error if the fetch fails.
func GetStructValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.StructValue, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetStructValue(), nil
}
