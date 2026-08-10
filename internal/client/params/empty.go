package params

import (
	"context"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles empty param creation in Terraform
// Read: Fetches current empty param state from device
// Update: Modifies existing empty params on device
// Delete: Removes empty params from device

// SetEmptyValue sets an empty/void parameter value via gRPC.
// This is typically used for trigger commands or reset operations.
func SetEmptyValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) error {
	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   oid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}
