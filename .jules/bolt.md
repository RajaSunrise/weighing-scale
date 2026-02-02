## 2025-02-18 - Unbounded Report Fetching
**Learning:** The `ShowReports` handler fetches *all* records matching a date range into memory without pagination. While I optimized the memory footprint by selecting only necessary columns, the lack of pagination remains a scalability risk if the dataset grows large.
**Action:** Future optimizations should implement keyset pagination or server-side datatables for the reports page, likely requiring frontend changes.
