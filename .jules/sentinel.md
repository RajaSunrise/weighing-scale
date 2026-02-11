## 2025-05-15 – Resource-intensive OCR (Tesseract) lacks concurrency control
**Vulnerability pattern:** Unbounded spawning of external processes via HTTP handlers creates a Remote Denial of Service (RDoS) vector.
**Learned constraint:** Always use a semaphore to limit concurrent execution of heavy external binaries (ffmpeg, tesseract, etc.).
**Prevention:** Implement a global or package-level weighted semaphore and wrap the process execution logic in a `TryAcquire` or blocking `Acquire` with timeout.
