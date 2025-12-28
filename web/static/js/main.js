document.addEventListener('DOMContentLoaded', () => {
    console.log("StoneWeigh UI Loaded");

    // Skip SSE initialization on weighing page since it has its own implementation
    if (window.location.pathname !== '/weighing') {
        // SSE Connection to Backend Stream
        // We use the protected API route. The browser cookies will handle auth.
        const evtSource = new EventSource("/api/scales/stream");

        evtSource.onmessage = function(event) {
            try {
                const data = JSON.parse(event.data);
                // data format: { scale_id: 1, weight: 12345, connected: true }

                // Dynamic ID lookup
                const displayEl = document.getElementById(`weight-display-${data.scale_id}`);
                const statusEl = document.getElementById(`status-scale-${data.scale_id}`);

                if (displayEl) {
                    // Update Weight
                    displayEl.innerText = data.weight.toFixed(0).padStart(5, '0');

                    // Update Status
                    if(statusEl) {
                        if (data.connected) {
                            statusEl.innerText = "TERHUBUNG";
                            statusEl.className = "px-2 py-1 rounded bg-success/10 text-success text-xs font-bold";
                        } else {
                            statusEl.innerText = "TERPUTUS";
                            statusEl.className = "px-2 py-1 rounded bg-red-500/10 text-red-500 text-xs font-bold";
                        }
                    }
                }

                // If we are currently "weighing" on this scale, update the form form values too
                const activeScaleId = document.getElementById('active-scale-id')?.value;
                if (activeScaleId && parseInt(activeScaleId) === data.scale_id) {
                    const grossEl = document.getElementById('val-gross');
                    const netEl = document.getElementById('val-net');
                    const tareEl = document.getElementById('val-tare');

                    if (grossEl && netEl) {
                        const gross = data.weight;
                        // Try to parse existing tare or default to 0
                        let tare = 0;
                        if(tareEl) {
                            tare = parseFloat(tareEl.innerText.replace(' kg', '')) || 0;
                        }

                        const net = gross - tare;

                        grossEl.innerText = `${gross} kg`;
                        netEl.innerText = `${net} kg`;
                    }
                }

            } catch (e) {
                console.error("Error parsing scale data", e);
            }
        };

        evtSource.onerror = function(err) {
            console.error("EventSource failed:", err);
        };
    }

    // Capture Button Logic
    window.captureWeight = function(scaleId) {
        const btn = event.currentTarget || document.activeElement;

        // Prevent recursive calls if triggered programmatically
        if (btn && btn.classList.contains('processing')) return;

        // Visual feedback if clicked
        if (btn && btn.tagName === 'BUTTON') {
             const originalContent = btn.innerHTML;
             btn.innerHTML = `<div class="loader border-white/20 border-t-white w-5 h-5"></div> Memproses...`;
             btn.classList.add('processing');
             btn.disabled = true;

             // Reset after short delay just for visual effect if ANPR is fast
             setTimeout(() => {
                 btn.innerHTML = originalContent;
                 btn.classList.remove('processing');
                 btn.disabled = false;
             }, 1000);
        }

        // Trigger ANPR
        // In real world, pass scale_id to select correct camera
        fetch(`/api/anpr/trigger?scale_id=${scaleId}`, { method: 'POST' })
            .then(res => res.json())
            .then(data => {
                const anprEl = document.getElementById('anpr-result');
                const plateInput = document.getElementById('plate_no');

                if(data.status === 'success' || data.status === 'simulated') {
                    if(anprEl) {
                        anprEl.innerText = data.plate;
                        anprEl.classList.remove('bg-primary');
                        anprEl.classList.add('bg-success');
                    }
                    if(plateInput) {
                        plateInput.value = data.plate;
                    }
                }
            })
            .catch(err => console.error(err));
    };
});
