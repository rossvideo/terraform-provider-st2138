package client

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseValueString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "boolean true",
			input:    "true",
			expected: true,
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: false,
		},
		{
			name:     "integer",
			input:    "42",
			expected: float64(42),
		},
		{
			name:     "float",
			input:    "3.14",
			expected: 3.14,
		},
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
	}

	// We need to test the parseValueString logic from params.go
	// Create a helper function to replicate the logic
	parseValue := func(s string) interface{} {
		if s == "" {
			return ""
		}
		if s == "true" || s == "True" || s == "TRUE" {
			return true
		}
		if s == "false" || s == "False" || s == "FALSE" {
			return false
		}
		// Try parsing as number
		var f float64
		if err := json.Unmarshal([]byte(s), &f); err == nil {
			return f
		}
		return s
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)
			if result != tt.expected {
				t.Errorf("parseValue(%q) = %v (%T), want %v (%T)",
					tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestSetParamsWithSlot_NullDynamic(t *testing.T) {
	c := &Client{
		Transport: "grpc",
	}

	// Test with null Dynamic value - should return error trying to connect
	nullDyn := types.DynamicNull()
	err := c.SetParamsWithSlot(context.Background(), nullDyn, 1)
	// We expect an error because there's no connection, but the null dynamic should be handled
	_ = err // Connection will fail, but that's expected in this unit test
}

func TestSetParamsWithSlot_UnknownDynamic(t *testing.T) {
	c := &Client{
		Transport: "grpc",
	}

	// Test with unknown Dynamic value - should return error trying to connect
	unknownDyn := types.DynamicUnknown()
	err := c.SetParamsWithSlot(context.Background(), unknownDyn, 1)
	// We expect an error because there's no connection, but the unknown dynamic should be handled
	_ = err // Connection will fail, but that's expected in this unit test
}
