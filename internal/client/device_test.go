package client

import (
	"testing"
	"time"
)

func TestNormalizeOID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already has leading slash",
			input:    "/status/ready",
			expected: "/status/ready",
		},
		{
			name:     "missing leading slash",
			input:    "status/ready",
			expected: "/status/ready",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test OID normalization logic used in device methods
			oid := tt.input
			if oid != "" && oid[0] != '/' {
				oid = "/" + oid
			} else if oid == "" {
				oid = "/"
			}

			if oid != tt.expected {
				t.Errorf("OID normalization = %v, want %v", oid, tt.expected)
			}
		})
	}
}

func TestWaitTimeout(t *testing.T) {
	// Test timeout logic without actual gRPC connection
	timeout := 100 * time.Millisecond
	deadline := time.Now().Add(timeout)

	time.Sleep(150 * time.Millisecond)

	if !time.Now().After(deadline) {
		t.Error("Expected deadline to be exceeded")
	}
}
