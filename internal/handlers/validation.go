package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

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
// It rejects control characters and HTML tags (< >) to prevent XSS
func validateSafeString(value string, fieldName string) error {
	// Check for control characters
	if strings.ContainsAny(value, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f") {
		return fmt.Errorf("%s contains invalid control characters", fieldName)
	}
	// Check for HTML injection characters
	if strings.ContainsAny(value, "<>") {
		return fmt.Errorf("%s contains invalid characters (< or >)", fieldName)
	}
	return nil
}

// validateUsername ensures the username is alphanumeric and starts with a letter
func validateUsername(value string) error {
	if !usernameRegex.MatchString(value) {
		return fmt.Errorf("username must start with a letter and contain only letters, numbers, and underscores")
	}
	return nil
}

// validateRole ensures the role is one of the allowed values
func validateRole(value string) error {
	if value != "admin" && value != "operator" {
		return fmt.Errorf("invalid role")
	}
	return nil
}
