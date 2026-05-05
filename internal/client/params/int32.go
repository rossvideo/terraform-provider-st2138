package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles int32 param creation in Terraform
// Read: Fetches current int32 param state from device
// Update: Modifies existing int32 params on device
// Delete: Removes int32 params from device

// SetInt32Value sets an int32 parameter value via gRPC.
// The OID is normalized to ensure it starts with '/'.
func SetInt32Value(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, value int32) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: value}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetInt32Value fetches an int32 parameter value via gRPC.
// Returns the int32 value or an error if the fetch fails.
func GetInt32Value(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (int32, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return 0, err
	}
	return val.GetInt32Value(), nil
}
