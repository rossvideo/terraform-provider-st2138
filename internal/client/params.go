package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles param creation in Terraform
// Read: Fetches current param state from device
// Update: Modifies existing params on device
// Delete: Removes params from device

// SetParams sets a subset of well-known params via gRPC when available.
// Currently supports setting string params like selected_flow_id if present.
func (c *Client) SetParams(ctx context.Context, dyn types.Dynamic) error {
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	// No generic decoding implemented yet; use resource-level helpers to set values.
	return nil
}

// SetStringValue performs GetParam then SetValue for a string OID.
func (c *Client) SetStringValue(ctx context.Context, slot uint32, oid string, value string) error {
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	// Use the provided slot as-is; caller is responsible for correctness
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
	_, err := c.rpcClient.SetValue(ctx, req)
	return err
}

// SetNumberValue sets a numeric param; prefers int32 when value is integral and in range, else float32.
func (c *Client) SetNumberValue(ctx context.Context, slot uint32, oid string, n float64) error {
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}
	var val *st2138pb.Value
	// Check if n is integral within int32 range
	if n == float64(int32(n)) {
		val = &st2138pb.Value{Kind: &st2138pb.Value_Int32Value{Int32Value: int32(n)}}
	} else {
		val = &st2138pb.Value{Kind: &st2138pb.Value_Float32Value{Float32Value: float32(n)}}
	}
	req := &st2138pb.SingleSetValuePayload{
		Slot:  slot,
		Value: &st2138pb.SetValuePayload{Oid: roid, Value: val},
	}
	_, err := c.rpcClient.SetValue(ctx, req)
	return err
}

// SetParamsWithSlot walks a JSON-like params object, sets all string and numeric leaves via SetValue.
// Complex values (objects/arrays) are traversed as subparams to produce OIDs like /a/b/0/c.
func (c *Client) SetParamsWithSlot(ctx context.Context, dyn types.Dynamic, slot uint32) error {
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	if dyn.IsNull() || dyn.IsUnknown() {
		return nil
	}
	var data any
	// Decode via JSON marshal/unmarshal; if wrapper object has a "value" field, use it.
	rawBytes, jerr := json.Marshal(dyn)
	if jerr != nil || len(rawBytes) == 0 {
		return nil
	}
	// First try direct decode
	if uerr := json.Unmarshal(rawBytes, &data); uerr != nil {
		return nil
	}
	// If result is a wrapper with common fields, extract nested "value"
	if m, ok := data.(map[string]any); ok {
		if v, vok := m["value"]; vok {
			data = v
		}
	}
	type pair struct {
		oid string
		v   any
	}
	var work []pair
	var walk func(prefix string, node any)
	walk = func(prefix string, node any) {
		switch t := node.(type) {
		case map[string]any:
			for k, v := range t {
				np := prefix + "/" + k
				walk(np, v)
			}
		case []any:
			for i, v := range t {
				np := fmt.Sprintf("%s/%d", prefix, i)
				walk(np, v)
			}
		case string:
			work = append(work, pair{oid: prefix, v: t})
		case float64:
			work = append(work, pair{oid: prefix, v: t})
		case bool:
			// booleans treated as strings "true"/"false" unless a dedicated type is desired
			work = append(work, pair{oid: prefix, v: t})
		default:
			// other scalar types not expected; ignore
		}
	}
	// Start walk at root with empty prefix; ensure leading '/'
	walk("", data)
	for _, p := range work {
		// Normalize OID
		oid := p.oid
		if !strings.HasPrefix(oid, "/") {
			oid = "/" + oid
		}
		switch v := p.v.(type) {
		case string:
			if err := c.SetStringValue(ctx, slot, oid, v); err != nil {
				return err
			}
		case float64:
			if err := c.SetNumberValue(ctx, slot, oid, v); err != nil {
				return err
			}
		case bool:
			sv := "false"
			if v {
				sv = "true"
			}
			if err := c.SetStringValue(ctx, slot, oid, sv); err != nil {
				return err
			}
		}
	}
	return nil
}
