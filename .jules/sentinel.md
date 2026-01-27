## 2026-01-27 - Missing Authorization in API Handlers
**Vulnerability:** IDOR in SaveTransaction allowed operators to access unassigned stations.
**Learning:** Middleware `AuthRequired` checks authentication but not specific resource access. Handlers must explicitly check `UserStationAssignment` for operators.
**Prevention:** Always verify ownership/assignment of the resource ID in the request payload against the session user ID, unless role is Admin.
