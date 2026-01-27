## 2026-01-27 – ANPR Trigger endpoint IDOR via dual parameter resolution
**Vulnerability pattern:** Endpoint accepts multiple optional ID parameters (`camera_id` or `scale_id`) but failed to enforce authorization on the resolved target station.
**Learned constraint:** When an endpoint can be triggered by multiple resource identifiers, authorization must be performed on the *resolved* common parent resource (Station) before action.
**Prevention:** Resolve target `WeighingStationID` first, then enforce `UserStationAssignment` check on that ID.

## 2025-05-20 – Sensitive token accepted via query parameter in Remote Scale API
**Vulnerability pattern:** Accepting authentication tokens via URL parameters causes leakage in logs and history
**Learned constraint:** Never read authentication tokens from `c.Query()`
**Prevention:** Strictly enforce Header-based authentication (e.g., `X-Scale-Token`) for all API endpoints
