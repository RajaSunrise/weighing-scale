## 2026-01-27 - GORM Foreign Key Indexing
**Learning:** GORM's `gorm.Model` and standard struct fields do not automatically create indexes for foreign keys (e.g., `UserID`, `StationID`). This results in O(N) table scans for frequent authorization lookups.
**Action:** Always explicitly verify and add `gorm:"index"` or composite indexes to foreign keys involved in hot-path queries (like permission checks in loop/stream handlers).
