package device

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDeviceResource_Metadata(t *testing.T) {
	r := &deviceResource{}
	req := resource.MetadataRequest{
		ProviderTypeName: "st2138",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "st2138_device" {
		t.Errorf("Metadata() TypeName = %s, want st2138_device", resp.TypeName)
	}
}

func TestDeviceResource_Schema(t *testing.T) {
	r := &deviceResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Schema.Description == "" {
		t.Error("Schema() Description should not be empty")
	}

	// Check that required attributes exist
	attrs := resp.Schema.Attributes
	if _, ok := attrs["id"]; !ok {
		t.Error("Schema should include 'id' attribute")
	}
	if _, ok := attrs["slot"]; !ok {
		t.Error("Schema should include 'slot' attribute")
	}
	if _, ok := attrs["name"]; !ok {
		t.Error("Schema should include 'name' attribute")
	}
}

func TestDeviceResource_ResolveDevicesDir(t *testing.T) {
	r := &deviceResource{}

	// Test with current directory (should always exist)
	result := r.resolveDevicesDir()
	if result == "" {
		t.Error("resolveDevicesDir() should return a non-empty path")
	}
}

func TestDeviceResource_PathExists_Extended(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "current directory",
			path: ".",
			want: true,
		},
		{
			name: "nonexistent directory",
			path: "/nonexistent/path/12345",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.pathExists(tt.path)
			if got != tt.want {
				t.Errorf("pathExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDeviceResource_CommandExists(t *testing.T) {
	r := &deviceResource{}

	// Test with a command that should exist on most systems
	if r.commandExists("ls") != true && r.commandExists("dir") != true {
		t.Error("commandExists should find 'ls' or 'dir'")
	}

	// Test with a nonexistent command
	if r.commandExists("nonexistentcommand12345") {
		t.Error("commandExists should return false for nonexistent command")
	}
}

func TestDeviceResource_DockerHost(t *testing.T) {
	r := &deviceResource{}

	result := r.dockerHost()

	// Result should be either localhost or host.docker.internal
	if result != "localhost" && result != "host.docker.internal" {
		t.Errorf("dockerHost() = %s, want localhost or host.docker.internal", result)
	}

	// If /.dockerenv exists, should return host.docker.internal
	if _, err := os.Stat("/.dockerenv"); err == nil {
		if result != "host.docker.internal" {
			t.Errorf("dockerHost() = %s, want host.docker.internal when in container", result)
		}
	} else {
		if result != "localhost" {
			t.Errorf("dockerHost() = %s, want localhost when not in container", result)
		}
	}
}

func TestDeviceResource_SelectHostPortForInternal_EdgeCases(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name     string
		ports    []string
		internal int
		want     int
	}{
		{
			name:     "multiple ports, first match",
			ports:    []string{"7254:6254", "8000:8080"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "multiple ports, second match",
			ports:    []string{"8000:8080", "7254:6254"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "with whitespace",
			ports:    []string{"  7254:6254  "},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "empty string in array",
			ports:    []string{"", "7254:6254"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "invalid format",
			ports:    []string{"invalid"},
			internal: 6254,
			want:     0,
		},
		{
			name:     "single port (invalid)",
			ports:    []string{"7254"},
			internal: 6254,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.selectHostPortForInternal(tt.ports, tt.internal)
			if got != tt.want {
				t.Errorf("selectHostPortForInternal() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDeviceResource_ParseValueString_AllTypes(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name      string
		input     string
		wantType  string
		wantValue interface{}
	}{
		{
			name:      "boolean true lowercase",
			input:     "true",
			wantType:  "bool",
			wantValue: true,
		},
		{
			name:      "boolean false lowercase",
			input:     "false",
			wantType:  "bool",
			wantValue: false,
		},
		{
			name:      "boolean TRUE uppercase",
			input:     "TRUE",
			wantType:  "bool",
			wantValue: true,
		},
		{
			name:      "boolean FALSE uppercase",
			input:     "FALSE",
			wantType:  "bool",
			wantValue: false,
		},
		{
			name:      "integer zero",
			input:     "0",
			wantType:  "float64",
			wantValue: float64(0),
		},
		{
			name:      "positive integer",
			input:     "123",
			wantType:  "float64",
			wantValue: float64(123),
		},
		{
			name:      "negative integer",
			input:     "-456",
			wantType:  "float64",
			wantValue: float64(-456),
		},
		{
			name:      "float with decimal",
			input:     "3.14159",
			wantType:  "float64",
			wantValue: 3.14159,
		},
		{
			name:      "negative float",
			input:     "-2.5",
			wantType:  "float64",
			wantValue: -2.5,
		},
		{
			name:      "scientific notation",
			input:     "1.5e10",
			wantType:  "float64",
			wantValue: 1.5e10,
		},
		{
			name:      "regular string",
			input:     "hello world",
			wantType:  "string",
			wantValue: "hello world",
		},
		{
			name:      "empty string",
			input:     "",
			wantType:  "string",
			wantValue: "",
		},
		{
			name:      "string that looks like bool but isn't",
			input:     "truthy",
			wantType:  "string",
			wantValue: "truthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.parseValueString(tt.input)

			var gotType string
			switch result.(type) {
			case bool:
				gotType = "bool"
			case float64:
				gotType = "float64"
			case string:
				gotType = "string"
			default:
				gotType = "unknown"
			}

			if gotType != tt.wantType {
				t.Errorf("parseValueString(%q) type = %s, want %s", tt.input, gotType, tt.wantType)
			}

			if result != tt.wantValue {
				t.Errorf("parseValueString(%q) = %v, want %v", tt.input, result, tt.wantValue)
			}
		})
	}
}

func TestDeviceResource_GetContainerID_EmptyCases(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab", "\t"},
		{"newline", "\n"},
		{"mixed whitespace", " \t\n "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.getContainerID(tt.input)
			if result != "" {
				t.Errorf("getContainerID(%q) = %q, want empty string", tt.input, result)
			}
		})
	}
}
