package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles int32 array param creation in Terraform
// Read: Fetches current int32 array param state from device
// Update: Modifies existing int32 array params on device
// Delete: Removes int32 array params from device

// SetInt32ArrayValue sets an int32 array parameter value via gRPC.
// The OID is normalized to ensure it starts with '/'.
func SetInt32ArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, intList *st2138pb.Int32List) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_Int32ArrayValues{Int32ArrayValues: intList}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetInt32ArrayValue fetches an int32 array parameter value via gRPC.
// Returns the int32 array or an error if the fetch fails.
func GetInt32ArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.Int32List, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetInt32ArrayValues(), nil
}
