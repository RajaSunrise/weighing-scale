## 2026-01-31 - Optimized Report Aggregation
**Learning:** Found a pattern where handlers fetched full datasets (e.g., for reporting) but still executed separate SQL aggregation queries (like SUM/COUNT) for summary statistics.
**Action:** When the full dataset is already in memory, perform aggregations (SUM, COUNT) in Go to eliminate redundant database roundtrips.
