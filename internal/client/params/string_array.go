package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles string array param creation in Terraform
// Read: Fetches current string array param state from device
// Update: Modifies existing string array params on device
// Delete: Removes string array params from device

// SetStringArrayValue sets a string array parameter value via gRPC.
// The OID is normalized to ensure it starts with '/'.
func SetStringArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, stringList *st2138pb.StringList) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_StringArrayValues{StringArrayValues: stringList}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetStringArrayValue fetches a string array parameter value via gRPC.
// Returns the string array or an error if the fetch fails.
func GetStringArrayValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.StringList, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetStringArrayValues(), nil
}
