package handlers

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// validateLength checks if a string's length is within the specified range.
// It uses utf8.RuneCountInString to count characters correctly.
func validateLength(value string, fieldName string, min int, max int) error {
	l := utf8.RuneCountInString(value)
	if l < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if l > max {
		return fmt.Errorf("%s must be at most %d characters", fieldName, max)
	}
	return nil
}

// validateSafeString ensures the string contains only safe characters
// This is a basic sanitization to prevent control characters or weird encodings
func validateSafeString(value string, fieldName string) error {
	if strings.ContainsAny(value, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f") {
		return fmt.Errorf("%s contains invalid control characters", fieldName)
	}
	return nil
}
