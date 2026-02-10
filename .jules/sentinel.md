# Sentinel Journal

## 2026-02-04 - Unvalidated RTSP URL Leading to SSRF
**Vulnerability:** The application allowed configuring `StationCamera` with `localhost`, `127.0.0.1`, or `169.254.x.x` as the RTSP URL. Since this URL is passed to `ffmpeg` on the server, it allowed Server-Side Request Forgery (SSRF) targeting internal services or cloud metadata.
**Learning:** Checking the URL scheme (e.g., ensuring it is `rtsp` or `http`) is NOT sufficient to prevent SSRF. Attackers can still target the loopback interface or link-local addresses using allowed schemes.
**Prevention:** Always parse the hostname and validate the IP address. Explicitly reject `IsLoopback()`, `IsLinkLocalUnicast()`, and `IsUnspecified()` (0.0.0.0) for any URL that will be accessed by the server.

## 2026-02-10 - Stored XSS in Log Viewer via Request Logger
**Vulnerability:** The system log viewer (`logs.html`) used `innerHTML` to render log lines. Since `RequestLogger` logged the raw request path, an attacker could inject `<script>` tags into the logs via a crafted URL path, which would execute when an admin viewed the logs.
**Learning:** `innerHTML` is inherently dangerous when rendering data from potentially untrusted sources, including server logs. Even "internal" data like logs can be poisoned by external inputs.
**Prevention:** Use `textContent` for rendering dynamic text in HTML. Sanitize all user-controlled data (like URL paths) before logging them to prevent log injection and secondary XSS.
