package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
)

// Create: Handles device creation in Terraform
// Read: Fetches current device state from backend
// Update: Modifies existing device on backend
// Delete: Removes device from backend

// RunStart triggers the device start via the SMPTE API using ExecuteCommand.
// commandOID should be a fully-qualified OID for the command; leading '/' will be added if missing.
// Supports both gRPC and HTTP transports.
func (c *Client) RunStart(ctx context.Context, slot uint32, commandOID string) error {
	oid := commandOID
	if !strings.HasPrefix(oid, "/") {
		oid = "/" + oid
	}

	emptyVal := &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
		stream, err := c.httpClient.ExecuteCommand(ctx, slot, oid, emptyVal, false)
		if err != nil {
			return err
		}
		if stream != nil {
			defer stream.Close()
		}
		return nil
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	payload := &st2138pb.ExecuteCommandPayload{
		Slot:    slot,
		Oid:     oid,
		Value:   emptyVal,
		Respond: false,
	}
	_, err := c.rpcClient.ExecuteCommand(ctx, payload)
	return err
}

// RunStop triggers a device stop via ExecuteCommand using the given command OID.
// Supports both gRPC and HTTP transports.
func (c *Client) RunStop(ctx context.Context, slot uint32, commandOID string) error {
	oid := commandOID
	if !strings.HasPrefix(oid, "/") {
		oid = "/" + oid
	}

	emptyVal := &st2138pb.Value{Kind: &st2138pb.Value_EmptyValue{EmptyValue: &st2138pb.Empty{}}}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return err
		}
		stream, err := c.httpClient.ExecuteCommand(ctx, slot, oid, emptyVal, false)
		if err != nil {
			return err
		}
		if stream != nil {
			defer stream.Close()
		}
		return nil
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}
	payload := &st2138pb.ExecuteCommandPayload{
		Slot:    slot,
		Oid:     oid,
		Value:   emptyVal,
		Respond: false,
	}
	_, err := c.rpcClient.ExecuteCommand(ctx, payload)
	return err
}

// WaitReady polls the given endpoint OID for the provided slot until the value equals readyValue or timeout elapses.
// Supports both gRPC and HTTP transports.
func (c *Client) WaitReady(ctx context.Context, slot uint32, endpoint string, readyValue string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		// Check context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Attempt to read current value
		val, err := c.GetStringValue(ctx, slot, endpoint)
		if err == nil && val == readyValue {
			return nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return err
			}
			return fmt.Errorf("timeout waiting for %s to equal %q (last=%q)", endpoint, readyValue, val)
		}
		time.Sleep(1 * time.Second)
	}
}

// WaitNotReady polls the given endpoint OID until the value differs from readyValue or timeout elapses.
// Supports both gRPC and HTTP transports.
func (c *Client) WaitNotReady(ctx context.Context, slot uint32, endpoint string, readyValue string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		val, err := c.GetStringValue(ctx, slot, endpoint)
		if err == nil && val != readyValue {
			return nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return err
			}
			return fmt.Errorf("timeout waiting for %s to differ from %q (last=%q)", endpoint, readyValue, val)
		}
		time.Sleep(1 * time.Second)
	}
}

// GetStringValue fetches a value for an OID and returns its string representation.
// If the underlying value is numeric or boolean, it is converted to a string.
// Supports both gRPC and HTTP transports.
func (c *Client) GetStringValue(ctx context.Context, slot uint32, oid string) (string, error) {
	roid := oid
	if !strings.HasPrefix(roid, "/") {
		roid = "/" + roid
	}

	// Use HTTP transport if configured
	if c.Transport == "http" || c.Transport == "https" || c.Transport == "rest" {
		if err := c.ensureHTTPConn(ctx); err != nil {
			return "", err
		}
		val, err := c.httpClient.GetValue(ctx, slot, roid)
		if err != nil {
			return "", err
		}
		// Prefer string if present; else coerce other scalar types
		if s := val.GetStringValue(); s != "" {
			return s, nil
		}
		if iv := val.GetInt32Value(); iv != 0 {
			return fmt.Sprintf("%d", iv), nil
		}
		if fv := val.GetFloat32Value(); fv != 0 {
			return fmt.Sprintf("%g", fv), nil
		}
		return "", nil
	}

	// Use gRPC transport (default)
	if err := c.ensureConn(ctx); err != nil {
		return "", err
	}
	req := &st2138pb.GetValuePayload{Slot: slot, Oid: roid}
	val, err := c.rpcClient.GetValue(ctx, req)
	if err != nil {
		return "", err
	}
	// Prefer string if present; else coerce other scalar types
	if s := val.GetStringValue(); s != "" {
		return s, nil
	}
	if iv := val.GetInt32Value(); iv != 0 {
		return fmt.Sprintf("%d", iv), nil
	}
	if fv := val.GetFloat32Value(); fv != 0 {
		return fmt.Sprintf("%g", fv), nil
	}
	// Default empty string if kind is empty
	return "", nil
}
