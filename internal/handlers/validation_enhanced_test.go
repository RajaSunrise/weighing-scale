package handlers

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateSafeString_Enhanced(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"Safe string", "Hello World", false},
		{"With number", "Vehicle 123", false},
		{"HTML tag start", "Hello <script>", true},
		{"HTML tag end", "Hello >", true},
		{"Newline LF", "Hello\nWorld", true},
		{"Newline CR", "Hello\rWorld", true},
		{"Ampersand", "A & B", true},
		{"Control char", "Hello \x01 World", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSafeString(tc.value, "TestField")
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
