## 2026-01-24 - [GORM Hook Validation for SSRF]
**Vulnerability:** The application accepts `RTSPURL` for cameras without validation. An attacker with admin access could inject `file://` or other malicious schemes, leading to SSRF or LFI when the URL is processed by FFmpeg/GoCV.
**Learning:** Admin interfaces are often trusted implicitly, but "Defense in Depth" requires validation at the data layer. FFmpeg is powerful and can read local files if not restricted.
**Prevention:** Use GORM `BeforeSave` hooks to enforce strict allowlists (e.g., `rtsp`, `http`) on URL fields at the model level. This protects all consumers of the data (API, Background Jobs, CLI) uniformly.
