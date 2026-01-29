## 2024-05-22 - Global Event Listeners Overhead
**Learning:** Initializing heavy resources like Server-Sent Events (SSE) in global JavaScript files (`main.js`) without page-specific checks causes unnecessary server load and network traffic on pages that don't need them.
**Action:** Always wrap heavy initializations in page-specific checks (e.g., checking `window.location.pathname` or presence of specific DOM elements) or move them to page-specific scripts.
