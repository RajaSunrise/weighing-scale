document.addEventListener('DOMContentLoaded', () => {
    console.log("StoneWeigh UI Loaded");

    const scaleElements = [1, 2, 3].map(id => ({
        display: document.getElementById(`weight-display-${id}`),
        status: document.getElementById(`status-scale-${id}`),
        container: document.getElementById(`weight-display-${id}`)?.closest('.relative')
    }));

    // SSE Connection to Backend Stream
    // We use the protected API route. The browser cookies will handle auth.
    const evtSource = new EventSource("/api/scales/stream");

    evtSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            // data format: { scale_id: 1, weight: 12345, connected: true }

            const idx = data.scale_id - 1;
            if (scaleElements[idx] && scaleElements[idx].display) {
                const el = scaleElements[idx];

                // Update Weight
                el.display.innerText = data.weight.toFixed(0).padStart(5, '0');

                // Update Status
                if(el.status) {
                    if (data.connected) {
                        el.status.innerText = "TERHUBUNG";
                        el.status.className = "px-2 py-1 rounded bg-success/10 text-success text-xs font-bold";
                    } else {
                        el.status.innerText = "TERPUTUS";
                        el.status.className = "px-2 py-1 rounded bg-red-500/10 text-red-500 text-xs font-bold";
                    }
                }
            }

            // If we are currently "weighing" on this scale, update the form form values too
            // For MVP, we assume Scale 1 is the active form scale
            if (data.scale_id === 1) {
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
        // EventSource auto-reconnects, but we might want to show UI state
    };

    // Capture Button Logic
    window.captureWeight = function(scaleId) {
        const btn = event.currentTarget;
        const originalContent = btn.innerHTML;
        btn.innerHTML = `<div class="loader border-white/20 border-t-white w-5 h-5"></div> Memproses...`;
        btn.disabled = true;

        // Trigger ANPR
        fetch('/api/anpr/trigger', { method: 'POST' })
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
            .catch(err => console.error(err))
            .finally(() => {
                btn.innerHTML = originalContent;
                btn.disabled = false;
            });
    };

    // Form Submission Logic
    const form = document.getElementById('weighing-form');
    if(form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);

            // Extract numbers from UI
            data.gross = parseFloat(document.getElementById('val-gross').innerText);
            data.tare = parseFloat(document.getElementById('val-tare').innerText);
            data.scale_id = 1; // Defaulting to scale 1 for now

            try {
                const res = await fetch('/api/transaction', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });

                const result = await res.json();
                if(res.ok) {
                    alert('Transaksi Berhasil! Tiket: ' + result.ticket);
                    // Open PDF
                    window.open('/' + result.invoice, '_blank');
                    // Reset
                    form.reset();
                    document.getElementById('anpr-result').innerText = "SCANNING...";
                    document.getElementById('anpr-result').className = "absolute -top-6 left-0 bg-primary text-white text-xs font-bold px-2 py-1 rounded";
                } else {
                    alert('Error: ' + result.error);
                }
            } catch (err) {
                alert('Connection Error');
            }
        });
    }
});
