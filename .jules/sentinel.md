# Sentinel Journal

## 2026-02-04 - Unvalidated RTSP URL Leading to SSRF
**Vulnerability:** The application allowed configuring `StationCamera` with `localhost`, `127.0.0.1`, or `169.254.x.x` as the RTSP URL. Since this URL is passed to `ffmpeg` on the server, it allowed Server-Side Request Forgery (SSRF) targeting internal services or cloud metadata.
**Learning:** Checking the URL scheme (e.g., ensuring it is `rtsp` or `http`) is NOT sufficient to prevent SSRF. Attackers can still target the loopback interface or link-local addresses using allowed schemes.
**Prevention:** Always parse the hostname and validate the IP address. Explicitly reject `IsLoopback()`, `IsLinkLocalUnicast()`, and `IsUnspecified()` (0.0.0.0) for any URL that will be accessed by the server.

## 2026-02-10 - Stored XSS in Log Viewer via Request Logger
**Vulnerability:** The system log viewer (`logs.html`) used `innerHTML` to render log lines. Since `RequestLogger` logged the raw request path, an attacker could inject `<script>` tags into the logs via a crafted URL path, which would execute when an admin viewed the logs.
**Learning:** `innerHTML` is inherently dangerous when rendering data from potentially untrusted sources, including server logs. Even "internal" data like logs can be poisoned by external inputs.
**Prevention:** Use `textContent` for rendering dynamic text in HTML. Sanitize all user-controlled data (like URL paths) before logging them to prevent log injection and secondary XSS.

## 2026-02-11 - CSRF Bypass via Loose Path Matching
**Vulnerability:** The CSRF middleware skipped protection for any path starting with `/api/external` using a loose slice check (`[:13]`). This allowed an attacker to bypass CSRF on sensitive endpoints by crafting paths like `/api/external_bypass` if such a route existed or could be exploited.
**Learning:** Prefix matching for security exceptions must be boundary-aware.
**Prevention:** Use exact path matching or ensure the prefix is followed by a path separator (e.g., `/api/external/`).

## 2026-02-11 - SQL LIKE Injection in Autocomplete
**Vulnerability:** User queries for vehicle plate numbers were directly concatenated into a `LIKE %...%` query without escaping special characters like `%` or `_`. This allowed users to perform broad searches that could disclose more data than intended or cause performance issues.
**Learning:** GORM's parameter binding protects against traditional SQL injection but does not escape special characters within a `LIKE` string.
**Prevention:** Explicitly escape `\`, `%`, and `_` in user-supplied strings before passing them to a `LIKE` query, and use the `ESCAPE` clause.

## 2026-02-11 – BOLA and Cache Poisoning in Dashboard and Report Charts
**Vulnerability pattern:** Broken Object Level Authorization (BOLA) through global aggregation and shared Redis cache keys.
**Learned constraint:** Always apply authorization scopes to aggregation queries (SUM, COUNT) and include user-specific identifiers in shared cache keys for non-admin users.
**Prevention:** Centralize authorization logic in GORM Scopes and enforce role-based cache key prefixing.
