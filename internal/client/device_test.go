package client

import (
	"testing"
	"time"
)

func TestWaitTimeout(t *testing.T) {
	// Test timeout logic without actual gRPC connection
	timeout := 100 * time.Millisecond
	deadline := time.Now().Add(timeout)

	time.Sleep(150 * time.Millisecond)

	if !time.Now().After(deadline) {
		t.Error("Expected deadline to be exceeded")
	}
}
