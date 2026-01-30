## 2026-01-30 - Input Validation Pattern
**Vulnerability:** Stored XSS via multiple input fields (Driver Name, Company, etc.)
**Learning:** The application relied on default Go template escaping for XSS protection, but some frontend components used `innerHTML` (e.g., `weighing.html` camera switcher and driver info), creating XSS sinks.
**Prevention:** Implemented a reusable `validateSafeString` helper in `internal/handlers/validation.go` that strictly rejects `<` and `>` characters. This is now applied to all user-input fields at the ingress point (API handlers), providing Defense in Depth.
