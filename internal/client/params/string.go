package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles string param creation in Terraform
// Read: Fetches current string param state from device
// Update: Modifies existing string params on device
// Delete: Removes string params from device

// SetStringValue sets a string parameter value via gRPC.
// The OID is normalized to ensure it starts with '/'.
func SetStringValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, value string) error {
	// Normalize OID: ensure it starts with '/'
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: value}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetStringValue fetches a string parameter value via gRPC.
// Returns the string value or an error if the fetch fails.
func GetStringValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (string, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return "", err
	}
	return val.GetStringValue(), nil
}
