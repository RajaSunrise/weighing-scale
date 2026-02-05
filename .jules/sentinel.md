# Sentinel Journal

## 2026-02-04 - Unvalidated RTSP URL Leading to SSRF
**Vulnerability:** The application allowed configuring `StationCamera` with `localhost`, `127.0.0.1`, or `169.254.x.x` as the RTSP URL. Since this URL is passed to `ffmpeg` on the server, it allowed Server-Side Request Forgery (SSRF) targeting internal services or cloud metadata.
**Learning:** Checking the URL scheme (e.g., ensuring it is `rtsp` or `http`) is NOT sufficient to prevent SSRF. Attackers can still target the loopback interface or link-local addresses using allowed schemes.
**Prevention:** Always parse the hostname and validate the IP address. Explicitly reject `IsLoopback()`, `IsLinkLocalUnicast()`, and `IsUnspecified()` (0.0.0.0) for any URL that will be accessed by the server.

## 2026-02-05 - SSRF via DNS Resolution Bypass
**Vulnerability:** The `validateRTSPURL` function relied on `net.ParseIP` which returns nil for hostnames, effectively skipping IP-based blocklists (Loopback, Link-Local) for any URL using a hostname (e.g., `http://metadata.google.internal/`).
**Learning:** `net.ParseIP` validates IP syntax but does NOT resolve hostnames. Validating URLs for SSRF requires explicitly resolving hostnames to IPs and checking the resolved addresses.
**Prevention:** Use `net.LookupIP` to resolve hostnames and validate all returned IPs against blocklists. Allow mocking of the lookup function for deterministic testing.
