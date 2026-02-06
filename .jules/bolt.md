## 2026-02-04 - Partial Column Selection
**Learning:** When using `.Select()` to optimize queries, ALWAYS include the Primary Key (ID) even if the current view doesn't explicitly use it. Future features or links often depend on IDs, and omitting them creates brittle code that breaks silently when templates evolve.
**Action:** Always add `id` to the `.Select()` list for GORM models.

## 2026-02-04 - Premature Optimization in Limited Queries
**Learning:** Optimizing queries with small, hard limits (e.g., `Limit(10)`) by selecting specific columns yields negligible performance gains but increases maintenance burden.
**Action:** Only apply column selection optimizations to queries that return large or unbounded datasets (like reports or exports).

## 2026-02-04 - Shared Cache Key Constants
**Learning:** Using magic strings for cache keys in multiple handlers (e.g., validation logic in master data vs read logic in reports) is error-prone. A single typo can lead to stale data bugs.
**Action:** Define cache keys as package-level constants in a central file (e.g., `handlers.go`) and reuse them across all related handlers.
