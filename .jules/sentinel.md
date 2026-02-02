## 2024-05-22 - Stored XSS in Hardware Configuration
**Vulnerability:** Stored XSS in `ScalePort` field of Weighing Station configuration.
**Learning:** "Technical" fields (ports, tokens, URLs) are often overlooked during input validation compared to "human" fields (names, descriptions). Even internal/admin-facing fields must be sanitized to prevent privilege escalation or lateral movement.
**Prevention:** Apply a "validate all inputs" policy. Use a shared validation helper (like `validateStationInput`) that explicitly checks every field in the struct, rather than checking fields individually in the handler which leads to omissions.
