## 2024-05-23 - SSRF in ProxyVideo

**Vulnerability:** Found a Server-Side Request Forgery (SSRF) vulnerability in the `/api/camera/stream` endpoint. The endpoint accepted a raw `url` query parameter which was passed directly to FFmpeg/GoCV, allowing an authenticated user to scan internal network ports or potentially access local files via the `file://` protocol.

**Learning:** This existed because the feature was designed for flexibility (allowing frontend to request streams) but lacked validation. The developer assumed that only valid camera URLs would be sent by the frontend, trusting the client input. Also, the `video_stream_gocv.go` file had a comment suggesting an intent to implement `camera_id` lookup but it wasn't implemented, falling back to raw URL usage.

**Prevention:** Never trust client input for URLs that are used in server-side requests or command executions. Always use an indirect reference map (ID -> URL) stored on the server side. If arbitrary URLs are absolutely required (rare), strict allowlisting of protocols, domains, and ports is mandatory. In this case, switching to `camera_id` lookups eliminated the risk entirely.
