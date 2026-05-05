package client

import (
	"testing"
)

func TestClient_Clone(t *testing.T) {
	original := &Client{
		Endpoint:   "localhost:6254",
		Transport:  "grpc",
		DevicesDir: "/devices",
	}

	clone := original.Clone()

	if clone.Endpoint != original.Endpoint {
		t.Errorf("Clone() Endpoint = %v, want %v", clone.Endpoint, original.Endpoint)
	}
	if clone.Transport != original.Transport {
		t.Errorf("Clone() Transport = %v, want %v", clone.Transport, original.Transport)
	}
	if clone.DevicesDir != original.DevicesDir {
		t.Errorf("Clone() DevicesDir = %v, want %v", clone.DevicesDir, original.DevicesDir)
	}
	if clone.conn != nil {
		t.Error("Clone() conn should be nil")
	}
	if clone.rpcClient != nil {
		t.Error("Clone() rpcClient should be nil")
	}
}

func TestClient_SetEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		initialEp    string
		newEp        string
		wantEndpoint string
	}{
		{
			name:         "set new endpoint",
			initialEp:    "localhost:6254",
			newEp:        "192.168.1.100:6254",
			wantEndpoint: "192.168.1.100:6254",
		},
		{
			name:         "set empty endpoint",
			initialEp:    "localhost:6254",
			newEp:        "",
			wantEndpoint: "localhost:6254",
		},
		{
			name:         "set same endpoint",
			initialEp:    "localhost:6254",
			newEp:        "localhost:6254",
			wantEndpoint: "localhost:6254",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Endpoint: tt.initialEp}
			c.SetEndpoint(tt.newEp)
			if c.Endpoint != tt.wantEndpoint {
				t.Errorf("SetEndpoint() endpoint = %v, want %v", c.Endpoint, tt.wantEndpoint)
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	c := &Client{
		Endpoint:  "localhost:6254",
		Transport: "grpc",
	}

	// Close should not panic even with nil connections
	c.Close()

	if c.conn != nil {
		t.Error("Close() conn should be nil")
	}
	if c.rpcClient != nil {
		t.Error("Close() rpcClient should be nil")
	}
}

func TestClient_ensureConn_InvalidTransport(t *testing.T) {
	c := &Client{
		Endpoint:  "localhost:6254",
		Transport: "http",
	}

	err := c.ensureConn(nil)
	if err == nil {
		t.Error("ensureConn() with invalid transport should return error")
	}
	if err.Error() != `transport "http" is not grpc` {
		t.Errorf("ensureConn() error = %v, want transport error", err)
	}
}
