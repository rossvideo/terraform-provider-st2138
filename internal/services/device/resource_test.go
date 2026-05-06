package device

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDeviceModel_Initialization(t *testing.T) {
	model := &deviceModel{
		Name: types.StringValue("test-device"),
	}

	if model.Name.ValueString() != "test-device" {
		t.Errorf("Name = %s, want test-device", model.Name.ValueString())
	}
}

func TestDeviceModel_NullValues(t *testing.T) {
	model := &deviceModel{
		ID:   types.StringNull(),
		Name: types.StringNull(),
	}

	if !model.ID.IsNull() {
		t.Error("ID should be null")
	}
	if !model.Name.IsNull() {
		t.Error("Name should be null")
	}
}

func TestDeviceModel_UnknownValues(t *testing.T) {
	model := &deviceModel{
		ID:   types.StringUnknown(),
		Name: types.StringUnknown(),
	}

	if !model.ID.IsUnknown() {
		t.Error("ID should be unknown")
	}
	if !model.Name.IsUnknown() {
		t.Error("Name should be unknown")
	}
}

func TestParamPairModel(t *testing.T) {
	pair := paramPairModel{
		Oid:   types.StringValue("/test/oid"),
		Value: types.StringValue("test-value"),
	}

	if pair.Oid.ValueString() != "/test/oid" {
		t.Errorf("Oid = %s, want /test/oid", pair.Oid.ValueString())
	}
	if pair.Value.ValueString() != "test-value" {
		t.Errorf("Value = %s, want test-value", pair.Value.ValueString())
	}
}

func TestDeviceStatusModel(t *testing.T) {
	status := &deviceStatusModel{
		Oid:        types.StringValue("/status/ready"),
		ReadyValue: types.StringValue("true"),
	}

	if status.Oid.ValueString() != "/status/ready" {
		t.Errorf("Oid = %s, want /status/ready", status.Oid.ValueString())
	}
	if status.ReadyValue.ValueString() != "true" {
		t.Errorf("ReadyValue = %s, want true", status.ReadyValue.ValueString())
	}
}

func TestDeviceResource_PathExists(t *testing.T) {
	r := &deviceResource{}

	// Test with current directory (should exist)
	if !r.pathExists(".") {
		t.Error("Current directory should exist")
	}

	// Test with non-existent path
	if r.pathExists("/nonexistent/path/12345") {
		t.Error("Non-existent path should return false")
	}
}

func TestDeviceResource_SelectHostPortForInternal(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name     string
		ports    []string
		internal int
		want     int
	}{
		{
			name:     "simple mapping",
			ports:    []string{"7254:6254"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "with ip address",
			ports:    []string{"127.0.0.1:7254:6254"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "with protocol",
			ports:    []string{"7254:6254/tcp"},
			internal: 6254,
			want:     7254,
		},
		{
			name:     "no match",
			ports:    []string{"8000:8080"},
			internal: 6254,
			want:     0,
		},
		{
			name:     "empty ports",
			ports:    []string{},
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

func TestDeviceResource_ParseValueString(t *testing.T) {
	r := &deviceResource{}

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "boolean true",
			input:    "true",
			wantType: "bool",
		},
		{
			name:     "boolean false",
			input:    "false",
			wantType: "bool",
		},
		{
			name:     "integer",
			input:    "42",
			wantType: "float64",
		},
		{
			name:     "float",
			input:    "3.14",
			wantType: "float64",
		},
		{
			name:     "string",
			input:    "hello",
			wantType: "string",
		},
		{
			name:     "empty",
			input:    "",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.parseValueString(tt.input)
			gotType := ""
			switch result.(type) {
			case bool:
				gotType = "bool"
			case float64:
				gotType = "float64"
			case string:
				gotType = "string"
			}
			if gotType != tt.wantType {
				t.Errorf("parseValueString(%q) type = %s, want %s", tt.input, gotType, tt.wantType)
			}
		})
	}
}

func TestDeviceResource_GetContainerID(t *testing.T) {
	r := &deviceResource{}

	// Test with empty name
	id := r.getContainerID("")
	if id != "" {
		t.Errorf("getContainerID(\"\") = %s, want empty string", id)
	}

	// Test with whitespace name
	id = r.getContainerID("   ")
	if id != "" {
		t.Errorf("getContainerID(\"   \") = %s, want empty string", id)
	}
}
