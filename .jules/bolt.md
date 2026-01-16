## 2024-05-22 - [Backend: In-Memory Aggregation Anti-Pattern]
**Learning:** The application fetches full datasets (albeit projected) to perform simple SUM aggregations based on date. This scales linearly with dataset size (O(N) memory/transfer) instead of constant O(1) transfer.
**Action:** Use SQL `GROUP BY` and `SUM` to push aggregation to the database engine.
