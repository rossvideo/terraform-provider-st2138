package params

import (
	"context"
	"strings"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles data param creation in Terraform
// Read: Fetches current data param state from device
// Update: Modifies existing data params on device
// Delete: Removes data params from device

// SetDataValue sets a data parameter value via gRPC.
// Data parameters are used to allow commands to accept file data.
// The OID is normalized to ensure it starts with '/'.
func SetDataValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string, dataPayload *st2138pb.DataPayload) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot: slot,
		Value: &st2138pb.SetValuePayload{
			Oid:   roid,
			Value: &st2138pb.Value{Kind: &st2138pb.Value_DataPayload{DataPayload: dataPayload}},
		},
	}
	_, err := client.SetValue(ctx, req)
	return err
}

// GetDataValue fetches a data parameter value via gRPC.
// Returns the data payload or an error if the fetch fails.
func GetDataValue(ctx context.Context, client st2138pb.CatenaServiceClient, slot uint32, oid string) (*st2138pb.DataPayload, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := client.GetValue(ctx, req)
	if err != nil {
		return nil, err
	}
	return val.GetDataPayload(), nil
}
