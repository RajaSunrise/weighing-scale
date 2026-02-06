package models

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSRF_DNSRebinding(t *testing.T) {
	// 1. Mock lookupIP to return 127.0.0.1 for a specific hostname
	originalLookupIP := lookupIP
	defer func() { lookupIP = originalLookupIP }()

	lookupIP = func(host string) ([]net.IP, error) {
		if host == "ssrf.local" {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}
		if host == "safe.com" {
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		}
		return originalLookupIP(host)
	}

	// 2. Test Case: malicious URL that looks safe (string-wise) but resolves to localhost
	// This hostname "ssrf.local" is NOT "localhost", so string check passes.
	// It is NOT an IP, so net.ParseIP returns nil.
	// It resolves to 127.0.0.1, so it SHOULD be blocked by IP check.
	maliciousURL := "rtsp://ssrf.local/stream"

	err := validateRTSPURL(maliciousURL)
	assert.Error(t, err, "Should reject URL that resolves to loopback IP")

	// 3. Test Case: Safe URL
	safeURL := "rtsp://safe.com/stream"
	err = validateRTSPURL(safeURL)
	assert.NoError(t, err, "Should accept safe URL")
}
