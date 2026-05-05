package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles float32 array param creation in Terraform
// Read: Fetches current float32 array param state from device
// Update: Modifies existing float32 array params on device
// Delete: Removes float32 array params from device

// SetFloat32ArrayValue sets a float32 array parameter value via gRPC.
// The OID is normalized to ensure it starts with '/'.
func SetFloat32ArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, floatList *st2138pb.Float32List) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_Float32ArrayValues{Float32ArrayValues: floatList}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetFloat32ArrayValue fetches a float32 array parameter value via gRPC.
// Returns the float32 array or an error if the fetch fails.
func GetFloat32ArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.Float32List, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetFloat32ArrayValues(), nil
}
