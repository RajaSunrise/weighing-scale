## 2025-05-22 - [Polling vs Event-Driven Streaming]
**Learning:** The video streaming handler used a 25ms polling loop (Ticker) to send frames to clients, regardless of whether a new frame was actually available from the camera. This caused high CPU usage (busy waking) and wasted bandwidth (sending duplicate frames).
**Action:** Replace polling loops with Go channels (`chan struct{}`) to signal availability. Clients should block and wait for the signal, ensuring they only wake up and transmit when a *new* frame is captured.
