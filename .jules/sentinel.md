# Sentinel Journal

## 2024-05-22 - Authorization Bypass in Video Proxy
**Vulnerability:** The `ProxyVideo` handler allowed any authenticated user to view any camera stream by guessing the `camera_id` or `station_id`, bypassing the intended station assignment restrictions.
**Learning:** `AuthRequired` middleware only authenticates the user; it does not authorize access to specific resources. Explicit authorization checks are needed for resource access.
**Prevention:** Implement resource-level authorization checks (e.g., verifying `UserStationAssignment`) in handlers that access specific resources, or use a centralized policy middleware.
