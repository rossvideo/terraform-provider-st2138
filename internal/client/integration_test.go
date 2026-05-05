package client_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/rossvideo/terraform-provider-st2138/internal/client"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
)

// Integration tests with a real mock gRPC server
// These test the full client stack without requiring a real ST2138 device

// mockServerState maintains state across gRPC calls
type mockServerState struct {
	values   map[string]*st2138pb.Value // key: "slot:oid"
	commands []string                    // track executed commands
}

type mockIntegrationServer struct {
	st2138pb.UnimplementedCatenaServiceServer
	state *mockServerState
}

func (s *mockIntegrationServer) SetValue(ctx context.Context, req *st2138pb.SingleSetValuePayload) (*st2138pb.Empty, error) {
	key := fmt.Sprintf("%d:%s", req.Slot, req.Value.Oid)
	s.state.values[key] = req.Value.Value
	return &st2138pb.Empty{}, nil
}

func (s *mockIntegrationServer) GetValue(ctx context.Context, req *st2138pb.GetValuePayload) (*st2138pb.Value, error) {
	key := fmt.Sprintf("%d:%s", req.Slot, req.Oid)
	if val, ok := s.state.values[key]; ok {
		return val, nil
	}
	// Return empty string as default
	return &st2138pb.Value{
		Kind: &st2138pb.Value_StringValue{StringValue: ""},
	}, nil
}

func (s *mockIntegrationServer) ExecuteCommand(req *st2138pb.ExecuteCommandPayload, stream st2138pb.CatenaService_ExecuteCommandServer) error {
	// Track command execution
	s.state.commands = append(s.state.commands, req.Oid)

	// For commands, set a status value
	if req.Oid == "/commands/start" {
		statusKey := fmt.Sprintf("%d:/status/state", req.Slot)
		s.state.values[statusKey] = &st2138pb.Value{
			Kind: &st2138pb.Value_StringValue{StringValue: "running"},
		}
	}

	return nil
}

// setupIntegrationServer starts a real gRPC server for testing
func setupIntegrationServer(t *testing.T) (string, *mockServerState, func()) {
	t.Helper()

	state := &mockServerState{
		values:   make(map[string]*st2138pb.Value),
		commands: make([]string, 0),
	}

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	st2138pb.RegisterCatenaServiceServer(grpcServer, &mockIntegrationServer{state: state})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	cleanup := func() {
		grpcServer.Stop()
		lis.Close()
	}

	return lis.Addr().String(), state, cleanup
}

// TestIntegration_BasicStringParameter tests setting and getting a string parameter
func TestIntegration_BasicStringParameter(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Set a string parameter
	err := c.SetStringValue(ctx, 1, "/config/device_name", "test-device")
	if err != nil {
		t.Fatalf("SetStringValue() error = %v", err)
	}

	// Verify it was stored in server state
	key := "1:/config/device_name"
	if val, ok := state.values[key]; !ok {
		t.Error("Value not stored in server")
	} else if val.GetStringValue() != "test-device" {
		t.Errorf("Stored value = %s, want test-device", val.GetStringValue())
	}

	// Get the string parameter back
	got, err := c.GetStringValue(ctx, 1, "/config/device_name")
	if err != nil {
		t.Fatalf("GetStringValue() error = %v", err)
	}

	if got != "test-device" {
		t.Errorf("GetStringValue() = %s, want test-device", got)
	}
}

// TestIntegration_NumericParameters tests integer and float parameters
func TestIntegration_NumericParameters(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Set an integer value
	err := c.SetNumberValue(ctx, 1, "/config/port", 6254.0)
	if err != nil {
		t.Fatalf("SetNumberValue(int) error = %v", err)
	}

	// Verify integer was stored as int32
	key := "1:/config/port"
	if val, ok := state.values[key]; !ok {
		t.Error("Integer value not stored")
	} else if val.GetInt32Value() != 6254 {
		t.Errorf("Stored int = %d, want 6254", val.GetInt32Value())
	}

	// Set a float value
	err = c.SetNumberValue(ctx, 1, "/config/threshold", 3.14159)
	if err != nil {
		t.Fatalf("SetNumberValue(float) error = %v", err)
	}

	// Verify float was stored as float32
	key = "1:/config/threshold"
	if val, ok := state.values[key]; !ok {
		t.Error("Float value not stored")
	} else if val.GetFloat32Value() == 0 {
		t.Error("Float value was not stored as float32")
	}
}

// TestIntegration_MultipleSlots tests managing multiple device slots
func TestIntegration_MultipleSlots(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Set values on different slots
	slots := []uint32{1, 2, 3}
	for _, slot := range slots {
		name := fmt.Sprintf("device-%d", slot)
		err := c.SetStringValue(ctx, slot, "/config/name", name)
		if err != nil {
			t.Fatalf("SetStringValue(slot %d) error = %v", slot, err)
		}
	}

	// Verify all slots have independent values
	for _, slot := range slots {
		expectedName := fmt.Sprintf("device-%d", slot)
		got, err := c.GetStringValue(ctx, slot, "/config/name")
		if err != nil {
			t.Fatalf("GetStringValue(slot %d) error = %v", slot, err)
		}
		if got != expectedName {
			t.Errorf("Slot %d: got %s, want %s", slot, got, expectedName)
		}
	}

	// Verify server has 3 separate entries
	if len(state.values) != 3 {
		t.Errorf("Server has %d values, want 3", len(state.values))
	}
}

// TestIntegration_CommandExecution tests start/stop commands
func TestIntegration_CommandExecution(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Execute start command
	err := c.RunStart(ctx, 1, "/commands/start")
	if err != nil {
		t.Fatalf("RunStart() error = %v", err)
	}

	// Give server time to process
	time.Sleep(50 * time.Millisecond)

	// Verify command was tracked
	if len(state.commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(state.commands))
	}
	if state.commands[0] != "/commands/start" {
		t.Errorf("Command = %s, want /commands/start", state.commands[0])
	}

	// Execute stop command
	err = c.RunStop(ctx, 1, "/commands/stop")
	if err != nil {
		t.Fatalf("RunStop() error = %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if len(state.commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(state.commands))
	}
}

// TestIntegration_OIDNormalization tests that OIDs are properly normalized
func TestIntegration_OIDNormalization(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Set value with and without leading slash - should be same
	err := c.SetStringValue(ctx, 1, "config/name", "test1")
	if err != nil {
		t.Fatalf("SetStringValue(no slash) error = %v", err)
	}

	err = c.SetStringValue(ctx, 1, "/config/name", "test2")
	if err != nil {
		t.Fatalf("SetStringValue(with slash) error = %v", err)
	}

	// Both should map to same key
	key := "1:/config/name"
	if val, ok := state.values[key]; !ok {
		t.Error("Value not found")
	} else if val.GetStringValue() != "test2" {
		t.Errorf("Value = %s, want test2 (second write should overwrite)", val.GetStringValue())
	}

	// Should only have one entry
	count := 0
	for k := range state.values {
		if k == "1:/config/name" || k == "1:config/name" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Found %d entries for config/name, want 1", count)
	}
}

// TestIntegration_ConcurrentAccess tests multiple goroutines accessing the client
func TestIntegration_ConcurrentAccess(t *testing.T) {
	endpoint, _, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()
	done := make(chan bool, 10)

	// Launch 10 concurrent goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("/test/concurrent_%d", id)
			value := fmt.Sprintf("value_%d", id)

			// Write
			err := c.SetStringValue(ctx, 1, key, value)
			if err != nil {
				t.Errorf("Goroutine %d: SetStringValue error = %v", id, err)
				return
			}

			// Read back
			got, err := c.GetStringValue(ctx, 1, key)
			if err != nil {
				t.Errorf("Goroutine %d: GetStringValue error = %v", id, err)
				return
			}

			if got != value {
				t.Errorf("Goroutine %d: got %s, want %s", id, got, value)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestIntegration_ConnectionReuse tests that client reuses connection
func TestIntegration_ConnectionReuse(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Make multiple calls - should reuse same connection
	for i := 0; i < 5; i++ {
		err := c.SetStringValue(ctx, 1, "/test/reuse", fmt.Sprintf("value%d", i))
		if err != nil {
			t.Fatalf("Call %d: error = %v", i, err)
		}
	}

	// All calls should have worked through same connection
	if len(state.values) != 1 {
		t.Errorf("Expected 1 final value, got %d", len(state.values))
	}
}

// TestIntegration_ClientClone tests cloning client for parallel resources
func TestIntegration_ClientClone(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	original := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer original.Close()

	clone := original.Clone()
	defer clone.Close()

	ctx := context.Background()

	// Use both clients concurrently
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		err := original.SetStringValue(ctx, 1, "/test/original", "from_original")
		if err != nil {
			t.Errorf("Original client error: %v", err)
		}
	}()

	go func() {
		defer func() { done <- true }()
		err := clone.SetStringValue(ctx, 2, "/test/clone", "from_clone")
		if err != nil {
			t.Errorf("Clone client error: %v", err)
		}
	}()

	<-done
	<-done

	// Both should have succeeded
	if len(state.values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(state.values))
	}
}

// TestIntegration_WaitReady tests status polling
func TestIntegration_WaitReady(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Set initial state to "starting"
	state.values["1:/status/state"] = &st2138pb.Value{
		Kind: &st2138pb.Value_StringValue{StringValue: "starting"},
	}

	// Launch goroutine to change state after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		state.values["1:/status/state"] = &st2138pb.Value{
			Kind: &st2138pb.Value_StringValue{StringValue: "ready"},
		}
	}()

	// Wait for ready state
	err := c.WaitReady(ctx, 1, "/status/state", "ready", 2*time.Second)
	if err != nil {
		t.Errorf("WaitReady() error = %v", err)
	}
}

// TestIntegration_WaitNotReady tests waiting for state change
func TestIntegration_WaitNotReady(t *testing.T) {
	endpoint, state, cleanup := setupIntegrationServer(t)
	defer cleanup()

	c := &client.Client{
		Endpoint:  endpoint,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Set initial state to "ready"
	state.values["1:/status/state"] = &st2138pb.Value{
		Kind: &st2138pb.Value_StringValue{StringValue: "ready"},
	}

	// Launch goroutine to change state after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		state.values["1:/status/state"] = &st2138pb.Value{
			Kind: &st2138pb.Value_StringValue{StringValue: "stopped"},
		}
	}()

	// Wait for not-ready state
	err := c.WaitNotReady(ctx, 1, "/status/state", "ready", 2*time.Second)
	if err != nil {
		t.Errorf("WaitNotReady() error = %v", err)
	}
}

// TestIntegration_EndpointChange tests changing endpoint
func TestIntegration_EndpointChange(t *testing.T) {
	endpoint1, _, cleanup1 := setupIntegrationServer(t)
	defer cleanup1()

	endpoint2, state2, cleanup2 := setupIntegrationServer(t)
	defer cleanup2()

	c := &client.Client{
		Endpoint:  endpoint1,
		Transport: "grpc",
	}
	defer c.Close()

	ctx := context.Background()

	// Write to first endpoint
	err := c.SetStringValue(ctx, 1, "/test/value", "endpoint1")
	if err != nil {
		t.Fatalf("SetStringValue(endpoint1) error = %v", err)
	}

	// Change endpoint
	c.SetEndpoint(endpoint2)

	// Write to second endpoint
	err = c.SetStringValue(ctx, 1, "/test/value", "endpoint2")
	if err != nil {
		t.Fatalf("SetStringValue(endpoint2) error = %v", err)
	}

	// Verify second endpoint received the value
	key := "1:/test/value"
	if val, ok := state2.values[key]; !ok {
		t.Error("Value not found on endpoint2")
	} else if val.GetStringValue() != "endpoint2" {
		t.Errorf("Value on endpoint2 = %s, want endpoint2", val.GetStringValue())
	}
}
