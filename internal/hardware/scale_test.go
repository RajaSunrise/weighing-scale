package hardware

import (
	"testing"
)

func TestParseWeight(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"ST,GS,+  12345 kg", 12345.0},
		{" 0.00 ", 0.0},
		{"- 50.5", -50.5},
		{"random text 100", 100.0},
		{"garbage", 0.0},
		{"12.34kg", 12.34},
	}

	for _, tt := range tests {
		result := parseWeight(tt.input)
		if result != tt.expected {
			t.Errorf("parseWeight(%q) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}
