document.addEventListener('DOMContentLoaded', () => {
    // Basic navigation highlighting is handled server-side via template logic,
    // but we can add client-side interactions here.

    console.log("StoneWeigh UI Loaded");

    // Mock WebSocket connection for Scale Data
    const scales = [
        { id: 1, weight: 0, connected: true },
        { id: 2, weight: 0, connected: false },
        { id: 3, weight: 0, connected: false }
    ];

    // Simulate Scale Data Updates
    setInterval(() => {
        if(document.getElementById('weight-display-1')) {
            // Jitter the weight slightly if connected to simulate real sensor noise
            if (scales[0].connected) {
                const noise = Math.floor(Math.random() * 5);
                // Keep it mostly 0 or simulate a truck coming on
                // For demo: oscillate between 0 and 25000 occasionally
                const base = Math.random() > 0.9 ? 24500 : 0;
                scales[0].weight = base + noise;

                updateScaleDisplay(1, scales[0].weight);
            }
        }
    }, 500);

    function updateScaleDisplay(id, weight) {
        const el = document.getElementById(`weight-display-${id}`);
        if(el) {
            el.innerText = weight.toString().padStart(5, '0');
        }
    }

    // Capture Button Logic
    window.captureWeight = function(scaleId) {
        const btn = event.currentTarget;
        const originalContent = btn.innerHTML;
        btn.innerHTML = `<div class="loader border-white/20 border-t-white w-5 h-5"></div> Processing...`;
        btn.disabled = true;

        // Simulate API Call to Capture & Analyze
        setTimeout(() => {
            // Mock Success
            btn.innerHTML = originalContent;
            btn.disabled = false;

            // Update Form
            const weight = document.getElementById(`weight-display-${scaleId}`).innerText;
            document.getElementById('val-gross').innerText = `${weight} kg`;
            document.getElementById('val-net').innerText = `${weight} kg`; // Assuming 0 tare for demo

            // Mock ANPR Result
            const plates = ["B 9821 XA", "BK 1122 YY", "D 8888 AA"];
            const randomPlate = plates[Math.floor(Math.random() * plates.length)];
            const anprEl = document.getElementById('anpr-result');
            const plateInput = document.getElementById('plate_no');

            if(anprEl) {
                anprEl.innerText = randomPlate;
                anprEl.classList.remove('bg-primary');
                anprEl.classList.add('bg-success');
            }
            if(plateInput) {
                plateInput.value = randomPlate;
            }

        }, 1500);
    };
});
