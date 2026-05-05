package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles float32 param creation in Terraform
// Read: Fetches current float32 param state from device
// Update: Modifies existing float32 params on device
// Delete: Removes float32 params from device

// SetFloat32Value sets a float32 parameter value via gRPC.
// The OID is normalized to ensure it starts with '/'.
func SetFloat32Value(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, value float32) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_Float32Value{Float32Value: value}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetFloat32Value fetches a float32 parameter value via gRPC.
// Returns the float32 value or an error if the fetch fails.
func GetFloat32Value(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (float32, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return 0, err
	}
	return val.GetFloat32Value(), nil
}
