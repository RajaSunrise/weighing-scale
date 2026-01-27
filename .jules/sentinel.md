## 2026-01-27 – ANPR Trigger endpoint IDOR via dual parameter resolution
**Vulnerability pattern:** Endpoint accepts multiple optional ID parameters (`camera_id` or `scale_id`) but failed to enforce authorization on the resolved target station.
**Learned constraint:** When an endpoint can be triggered by multiple resource identifiers, authorization must be performed on the *resolved* common parent resource (Station) before action.
**Prevention:** Resolve target `WeighingStationID` first, then enforce `UserStationAssignment` check on that ID.
