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

// SetParams sets a subset of well-known params via transport when available.
// Currently supports setting string params like selected_flow_id if present.
// Supports both gRPC and HTTP transports.
func (c *Client) SetParams(ctx context.Context, dyn types.Dynamic) error {
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
	} else {
		if err := c.ensureConn(ctx); err != nil {
			return err
		}
	}
	// No generic decoding implemented yet; use resource-level helpers to set values.
	return nil
}

// SetStringValue performs GetParam then SetValue for a string OID.
// Supports both gRPC and HTTP transports.
func (c *Client) SetStringValue(ctx context.Context, slot uint32, oid string, value string) error {
	// Normalize OID: ensure it starts with '/'
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	val := &st2138pb.Value{Kind: &st2138pb.Value_StringValue{StringValue: value}}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
		return c.httpClient.SetValue(ctx, slot, roid, val)
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot:  slot,
		Value: &st2138pb.SetValuePayload{Oid: roid, Value: val},
	}
	_, err := c.rpcClient.SetValue(ctx, req)
	return err
}

// SetNumberValue sets a numeric param; prefers int32 when value is integral and in range, else float32.
// Supports both gRPC and HTTP transports.
func (c *Client) SetNumberValue(ctx context.Context, slot uint32, oid string, n float64) error {
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

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
		return c.httpClient.SetValue(ctx, slot, roid, val)
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot:  slot,
		Value: &st2138pb.SetValuePayload{Oid: roid, Value: val},
	}
	_, err := c.rpcClient.SetValue(ctx, req)
	return err
}

// SetRawValue sends a fully-formed Catena value payload for the given OID.
// Supports both gRPC and HTTP transports.
func (c *Client) SetRawValue(ctx context.Context, slot uint32, oid string, value *st2138pb.Value) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
		return c.httpClient.SetValue(ctx, slot, roid, value)
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}

	req := &st2138pb.SingleSetValuePayload{
		Slot:  slot,
		Value: &st2138pb.SetValuePayload{Oid: roid, Value: value},
	}
	_, err := c.rpcClient.SetValue(ctx, req)
	return err
}

// GetParamDescriptor fetches the parameter descriptor for an OID.
// Supports both gRPC and HTTP transports.
func (c *Client) GetParamDescriptor(ctx context.Context, slot uint32, oid string) (*st2138pb.Param, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return nil, err
		}
		return c.httpClient.GetParam(ctx, slot, roid)
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return nil, err
	}
	resp, err := c.rpcClient.GetParam(ctx, &st2138pb.GetParamPayload{Slot: slot, Oid: roid})
	if err != nil {
		return nil, err
	}
	return resp.GetParam(), nil
}

// ExecuteCommand invokes a device command and drains the streaming response.
// value may be nil if the command takes no parameter.
// Supports both gRPC and HTTP transports.
func (c *Client) ExecuteCommand(ctx context.Context, slot uint32, oid string, value *st2138pb.Value) error {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
		stream, err := c.httpClient.ExecuteCommand(ctx, slot, roid, value, true)
		if err != nil {
			return fmt.Errorf("ExecuteCommand %s: %w", oid, err)
		}
		defer stream.Close()
		for {
			resp, recvErr := stream.RecvCommandResponse()
			if recvErr != nil {
				if recvErr.Error() == "EOF" {
					break
				}
				break
			}
			if ex := resp.GetException(); ex != nil {
				return fmt.Errorf("command %s exception: %s", oid, ex.GetDetails())
			}
		}
		return nil
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	stream, err := c.rpcClient.ExecuteCommand(ctx, &st2138pb.ExecuteCommandPayload{
		Slot:    slot,
		Oid:     roid,
		Value:   value,
		Respond: true,
	})
	if err != nil {
		return fmt.Errorf("ExecuteCommand %s: %w", oid, err)
	}
	for {
		resp, recvErr := stream.Recv()
		if recvErr != nil {
			if recvErr.Error() == "EOF" {
				break
			}
			break
		}
		if ex := resp.GetException(); ex != nil {
			return fmt.Errorf("command %s exception: %s", oid, ex.GetDetails())
		}
	}
	return nil
}

// GetRawValue fetches the current proto Value for the given OID.
// Supports both gRPC and HTTP transports.
func (c *Client) GetRawValue(ctx context.Context, slot uint32, oid string) (*st2138pb.Value, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return nil, err
		}
		return c.httpClient.GetValue(ctx, slot, roid)
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return nil, err
	}
	return c.rpcClient.GetValue(ctx, &st2138pb.GetValuePayload{Slot: slot, Oid: roid})
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
