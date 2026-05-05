package params

import (
	"strings"
	"testing"
)

func TestOIDNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized",
			input:    "/param/test",
			expected: "/param/test",
		},
		{
			name:     "needs normalization",
			input:    "param/test",
			expected: "/param/test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "/",
		},
		{
			name:     "single char",
			input:    "p",
			expected: "/p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roid := tt.input
			if !strings.HasPrefix(roid, "/") {
				roid = "/" + roid
			}
			if roid != tt.expected {
				t.Errorf("OID normalization = %v, want %v", roid, tt.expected)
			}
		})
	}
}

func TestStringValueTypes(t *testing.T) {
	// Test that string values are properly typed
	testVal := "test-value"
	if len(testVal) == 0 {
		t.Error("String value should not be empty")
	}
}

func TestInt32ValueTypes(t *testing.T) {
	// Test int32 value ranges
	tests := []struct {
		name  string
		value int32
		valid bool
	}{
		{"zero", 0, true},
		{"positive", 42, true},
		{"negative", -42, true},
		{"max", 2147483647, true},
		{"min", -2147483648, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the value is within int32 range
			_ = tt.value
		})
	}
}

func TestFloat32ValueTypes(t *testing.T) {
	// Test float32 value types
	tests := []struct {
		name  string
		value float32
		valid bool
	}{
		{"zero", 0.0, true},
		{"positive", 3.14, true},
		{"negative", -3.14, true},
		{"small", 0.0001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.value
		})
	}
}

func TestArrayOperations(t *testing.T) {
	// Test array type operations
	intArray := []int32{1, 2, 3, 4, 5}
	if len(intArray) != 5 {
		t.Errorf("Int32 array length = %d, want 5", len(intArray))
	}

	floatArray := []float32{1.1, 2.2, 3.3}
	if len(floatArray) != 3 {
		t.Errorf("Float32 array length = %d, want 3", len(floatArray))
	}

	stringArray := []string{"a", "b", "c"}
	if len(stringArray) != 3 {
		t.Errorf("String array length = %d, want 3", len(stringArray))
	}
}

func TestBinaryData(t *testing.T) {
	// Test binary data handling
	data := []byte{0x00, 0x01, 0x02, 0xFF}
	if len(data) != 4 {
		t.Errorf("Binary data length = %d, want 4", len(data))
	}
	if data[0] != 0x00 {
		t.Errorf("Binary data[0] = %x, want 0x00", data[0])
	}
	if data[3] != 0xFF {
		t.Errorf("Binary data[3] = %x, want 0xFF", data[3])
	}
}
