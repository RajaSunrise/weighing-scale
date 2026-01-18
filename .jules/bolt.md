## 2024-05-23 - Missing Index on WeighingRecord.WeighedAt
**Learning:** The `WeighingRecord` table lacked an index on `WeighedAt`, despite this field being used for all reporting queries (range filtering) and dashboard stats. This would cause O(N) performance on large datasets.
**Action:** Always check schema definitions for indexes on fields used in WHERE and ORDER BY clauses, especially for time-series data.
