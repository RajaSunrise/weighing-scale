
// Mock data and helper functions
const { useState, useEffect } = React;
const { createRoot } = ReactDOM;
// Since we are doing multi-page with Go, we might not strictly need React Router for top level nav,
// but the components use Link and useNavigate.
// We can use a HashRouter or just mock these if we want true multi-page.
// However, the user asked to "pecah" (split) the code.
// A hybrid approach: Use BrowserRouter but server routes match.
const { BrowserRouter, Routes, Route, Link, useNavigate, useLocation } = ReactRouterDOM;

// --- Navigation Helper Component ---
const MainNavigation = () => {
    // In a multi-page app, navigating to "/dashboard" should trigger a full page load
    // if we want to serve a different HTML file.
    // If we use React Router Link, it intercepts and does client-side nav.
    // To support the "Split into pages" requirement where Go serves different templates,
    // we should use normal <a href> or ensure Link causes a reload?
    // Actually, if we use React Router, we can just keep it as an SPA.
    // BUT the user said "split the code into how many pages".
    // I will implement this as a multi-page app where components are reused.
    // So "Link" should probably just be an <a> tag wrapper or we configure Router to refresh.

    // For simplicity in this hybrid mode:
    const handleNav = (path) => {
        window.location.href = path;
    };

    const [isOpen, setIsOpen] = useState(false);

    const screens = [
        { path: "/", name: "Login Portal" },
        { path: "/dashboard", name: "Main Dashboard" },
        { path: "/weighing-station", name: "Weighing Station (Operator)" },
        { path: "/report-dashboard", name: "Report Dashboard" },
        // Add others as needed
    ];

    return (
        <div className="fixed bottom-4 right-4 z-[9999]">
            {isOpen && (
                <div className="bg-white dark:bg-[#1a222d] border border-slate-200 dark:border-slate-700 rounded-lg shadow-2xl p-2 mb-2 flex flex-col gap-1 max-h-[80vh] overflow-y-auto w-64">
                    <div className="px-2 py-1 text-xs font-bold text-slate-500 uppercase">Navigation</div>
                    {screens.map((screen) => (
                        <button
                            key={screen.path}
                            onClick={() => {
                                handleNav(screen.path);
                                setIsOpen(false);
                            }}
                            className={`text-left px-3 py-2 rounded text-sm hover:bg-slate-100 dark:hover:bg-slate-700`}
                        >
                            {screen.name}
                        </button>
                    ))}
                </div>
            )}
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center justify-center size-12 rounded-full bg-primary text-white shadow-lg hover:bg-blue-600 transition-colors"
            >
                <span className="material-symbols-outlined">{isOpen ? 'close' : 'menu'}</span>
            </button>
        </div>
    );
};

// Screen: Login
const LoginScreen = () => {
    // We use window.location for navigation in this multi-page setup
    return (
        <div className="h-screen flex flex-col overflow-hidden bg-background-light dark:bg-background-dark">
            <header className="flex items-center justify-between whitespace-nowrap border-b border-solid border-gray-200 dark:border-border-dark px-6 lg:px-10 py-4 bg-white dark:bg-background-dark z-20 shadow-sm relative shrink-0">
                <div className="flex items-center gap-4">
                    <div className="size-10 bg-primary/10 rounded-lg flex items-center justify-center text-primary">
                        <span className="material-symbols-outlined text-[28px]">scale</span>
                    </div>
                    <div>
                        <h2 className="text-gray-900 dark:text-white text-lg font-bold leading-tight tracking-[-0.015em]">Sistem Penimbangan Terpadu</h2>
                        <p className="text-xs text-gray-500 dark:text-text-secondary font-medium">Monitoring & Manajemen Data Timbangan</p>
                    </div>
                </div>
                <div className="flex items-center gap-4">
                    <button className="hidden sm:flex h-10 px-4 items-center justify-center rounded-lg bg-gray-100 dark:bg-surface-dark text-gray-700 dark:text-white text-sm font-bold border border-transparent hover:border-gray-300 dark:hover:border-border-dark transition-colors">
                        <span className="material-symbols-outlined mr-2 text-[18px]">help</span>
                        <span>Bantuan</span>
                    </button>
                    <div className="h-8 w-[1px] bg-gray-200 dark:bg-border-dark hidden sm:block"></div>
                    <div className="flex items-center gap-2 text-gray-500 dark:text-text-secondary">
                        <span className="material-symbols-outlined text-[20px]">language</span>
                        <span className="text-sm font-semibold">ID</span>
                    </div>
                </div>
            </header>
            <main className="flex-1 flex overflow-hidden relative">
                <div className="flex-1 flex flex-col overflow-y-auto no-scrollbar relative z-10 w-full lg:w-1/2 lg:max-w-[640px] bg-white dark:bg-background-dark border-r border-gray-200 dark:border-border-dark shadow-2xl">
                    <div className="flex-1 flex flex-col justify-center px-6 sm:px-12 py-10">
                        <div className="max-w-[480px] w-full mx-auto">
                            <div className="mb-8">
                                <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">Portal Pengelola</h1>
                                <p className="text-gray-500 dark:text-text-secondary">Kelola akses timbangan, CCTV ANPR, dan pelaporan tiket digital dalam satu dashboard terintegrasi.</p>
                            </div>
                            <div className="mb-8 border-b border-gray-200 dark:border-border-dark">
                                <nav aria-label="Tabs" className="flex space-x-8">
                                    <a className="border-primary text-primary whitespace-nowrap py-4 px-1 border-b-2 font-bold text-sm flex items-center gap-2" href="#">
                                        <span className="material-symbols-outlined text-[20px]">login</span>
                                        Masuk
                                    </a>
                                    <a className="border-transparent text-gray-500 dark:text-text-secondary hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2" href="#">
                                        <span className="material-symbols-outlined text-[20px]">person_add</span>
                                        Daftar
                                    </a>
                                </nav>
                            </div>
                            <form className="flex flex-col gap-5" onSubmit={(e) => {
                                e.preventDefault();
                                // Mock API Call
                                fetch('/api/login', {
                                    method: 'POST',
                                    headers: {'Content-Type': 'application/json'},
                                    body: JSON.stringify({email: e.target.email.value, password: e.target.password.value})
                                })
                                .then(res => {
                                    if (res.ok) {
                                        window.location.href='/dashboard';
                                    } else {
                                        alert("Login failed");
                                    }
                                })
                                .catch(err => {
                                    console.error(err);
                                    // Fallback for prototype if API fails/offline
                                    window.location.href='/dashboard';
                                });
                            }}>
                                <div className="space-y-4">
                                    <div>
                                        <label className="block text-sm font-semibold text-gray-700 dark:text-white mb-2" htmlFor="email">Email atau ID Pengguna</label>
                                        <div className="relative">
                                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400 dark:text-text-secondary">
                                                <span className="material-symbols-outlined text-[20px]">mail</span>
                                            </div>
                                            <input className="block w-full pl-10 pr-3 py-3 border border-gray-300 dark:border-border-dark rounded-lg leading-5 bg-white dark:bg-surface-dark text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary sm:text-sm transition duration-150 ease-in-out" id="email" name="email" placeholder="admin@perusahaan.com" type="text"/>
                                        </div>
                                    </div>
                                    <div>
                                        <div className="flex items-center justify-between mb-2">
                                            <label className="block text-sm font-semibold text-gray-700 dark:text-white" htmlFor="password">Kata Sandi</label>
                                            <a className="text-sm font-bold text-primary hover:text-primary-hover" href="#">Lupa sandi?</a>
                                        </div>
                                        <div className="relative">
                                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400 dark:text-text-secondary">
                                                <span className="material-symbols-outlined text-[20px]">lock</span>
                                            </div>
                                            <input className="block w-full pl-10 pr-10 py-3 border border-gray-300 dark:border-border-dark rounded-lg leading-5 bg-white dark:bg-surface-dark text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary sm:text-sm transition duration-150 ease-in-out" id="password" name="password" placeholder="••••••••" type="password"/>
                                            <div className="absolute inset-y-0 right-0 pr-3 flex items-center cursor-pointer text-gray-400 dark:text-text-secondary hover:text-gray-600 dark:hover:text-white">
                                                <span className="material-symbols-outlined text-[20px]">visibility_off</span>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex items-start gap-3 p-3 rounded-lg bg-blue-50 dark:bg-primary/10 border border-blue-100 dark:border-primary/20">
                                    <span className="material-symbols-outlined text-primary text-[20px] mt-0.5">shield</span>
                                    <p className="text-xs text-gray-600 dark:text-blue-100">
                                        Sesi anda dilindungi dengan enkripsi end-to-end. Pastikan anda logout setelah selesai mengelola data timbangan.
                                    </p>
                                </div>
                                <button className="mt-2 w-full flex justify-center py-3.5 px-4 border border-transparent rounded-lg shadow-sm text-sm font-bold text-white bg-primary hover:bg-primary-hover focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary transition-all shadow-lg shadow-primary/25" type="submit">
                                    Masuk ke Dashboard
                                </button>
                            </form>
                            <div className="mt-8 pt-6 border-t border-gray-200 dark:border-border-dark flex flex-col sm:flex-row justify-between items-center gap-4 text-xs text-gray-500 dark:text-text-secondary">
                                <p>© 2024 WeighBridge Systems Inc.</p>
                                <div className="flex gap-4">
                                    <a className="hover:text-primary" href="#">Kebijakan Privasi</a>
                                    <a className="hover:text-primary" href="#">Syarat & Ketentuan</a>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="hidden lg:flex flex-1 relative bg-surface-dark overflow-hidden">
                     <div className="absolute inset-0 z-0">
                        <img alt="Industrial warehouse abstract view" className="w-full h-full object-cover opacity-40 mix-blend-overlay" src="https://lh3.googleusercontent.com/aida-public/AB6AXuAuLgNu4f4j4ToA3lf-0dKRu4LgoH0prSmJCwk0VdLQAcx7p5DvKPkMNWeW00TdDHKh66M7lfgBrndwVrrYJ6mu7s1ZvKOoRj3buet2GUk0LGtYh4ekUDBx2ZvPl8tWGI9aTFRVvdAXTzwSDYs8j-QEgZdV4CZwrFhlVw83YT6PrN65-TS65eGfnfKWAF-qGezBOAfJWCcI0XFW8P9x5Fdl6UkTOjLsdSyiiFRy2IPoQH1XLXRKbLEvGsHyd-R3fpoV9Ys3AyKb0i8"/>
                        <div className="absolute inset-0 bg-gradient-to-t from-background-dark via-background-dark/80 to-primary/20 mix-blend-multiply"></div>
                    </div>
                    <div className="relative z-10 flex flex-col justify-end p-12 w-full">
                        <div className="grid grid-cols-2 gap-4 mb-8">
                            <div className="bg-surface-dark/60 backdrop-blur-md border border-border-dark/50 p-4 rounded-xl">
                                <div className="flex items-center gap-3 mb-2">
                                    <div className="size-8 rounded-full bg-emerald-500/20 text-emerald-400 flex items-center justify-center">
                                        <span className="material-symbols-outlined text-[18px]">videocam</span>
                                    </div>
                                    <span className="text-xs font-bold uppercase tracking-wider text-text-secondary">CCTV ANPR</span>
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className="text-2xl font-bold text-white">Online</span>
                                    <div className="h-2 w-2 rounded-full bg-emerald-500 mb-2 animate-pulse"></div>
                                </div>
                                <p className="text-xs text-gray-400 mt-1">Deteksi plat otomatis aktif</p>
                            </div>
                            <div className="bg-surface-dark/60 backdrop-blur-md border border-border-dark/50 p-4 rounded-xl">
                                <div className="flex items-center gap-3 mb-2">
                                    <div className="size-8 rounded-full bg-primary/20 text-primary flex items-center justify-center">
                                        <span className="material-symbols-outlined text-[18px]">local_shipping</span>
                                    </div>
                                    <span className="text-xs font-bold uppercase tracking-wider text-text-secondary">Timbangan</span>
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className="text-2xl font-bold text-white">Kalibrasi</span>
                                    <span className="text-xs text-emerald-400 mb-1 font-medium">OK</span>
                                </div>
                                <p className="text-xs text-gray-400 mt-1">Status sensor stabil</p>
                            </div>
                        </div>
                        <div className="bg-gradient-to-br from-gray-900/90 to-gray-800/90 backdrop-blur-xl border border-border-dark p-8 rounded-2xl shadow-2xl max-w-lg">
                            <div className="flex items-start gap-4">
                                <div className="p-3 bg-primary rounded-lg shadow-lg shadow-primary/30">
                                    <span className="material-symbols-outlined text-white text-[28px]">dataset</span>
                                </div>
                                <div>
                                    <h3 className="text-xl font-bold text-white mb-2">Manajemen Data Terpusat</h3>
                                    <p className="text-gray-300 text-sm leading-relaxed">
                                        Pantau lalu lintas kendaraan tambang secara real-time. Sistem terintegrasi dengan PostgreSQL untuk penyimpanan data tiket timbangan yang aman, akurat, dan dapat diaudit.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};

// Screen: Dashboard
const DashboardScreen = () => {
    const [stats, setStats] = useState({total_count: 0, total_weight: 0, pending_count: 0});
    const [recentTx, setRecentTx] = useState([]);

    useEffect(() => {
        const fetchData = () => {
            fetch('/api/stats')
                .then(res => res.json())
                .then(data => setStats(data))
                .catch(console.error);

            fetch('/api/transactions')
                .then(res => res.json())
                .then(data => setRecentTx(data))
                .catch(console.error);
        };

        fetchData();
        const interval = setInterval(fetchData, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="flex h-screen w-full bg-background-light dark:bg-background-dark overflow-hidden">
            <aside className="w-20 lg:w-72 flex-shrink-0 flex flex-col justify-between border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-[#111822] transition-all duration-300">
                <div className="flex flex-col h-full p-4 gap-6">
                    <div className="flex items-center gap-3 px-2">
                        <div className="flex items-center justify-center size-10 rounded-xl bg-gradient-to-br from-primary to-blue-600 shadow-lg shadow-blue-500/20 text-white">
                            <span className="material-symbols-outlined text-2xl">scale</span>
                        </div>
                        <div className="flex flex-col hidden lg:flex">
                            <h1 className="text-base font-bold leading-none">StoneWeigh</h1>
                            <p className="text-text-secondary text-xs mt-1">Sistem Timbangan</p>
                        </div>
                    </div>
                    <nav className="flex flex-col gap-2 flex-1">
                        <a className="flex items-center gap-3 px-3 py-3 rounded-xl bg-primary/10 text-primary transition-colors" href="/dashboard">
                            <span className="material-symbols-outlined">dashboard</span>
                            <span className="text-sm font-bold hidden lg:block">Dashboard</span>
                        </a>
                        <a className="flex items-center gap-3 px-3 py-3 rounded-xl text-slate-500 dark:text-text-secondary hover:bg-slate-100 dark:hover:bg-card-hover transition-colors" href="/report-dashboard">
                            <span className="material-symbols-outlined">description</span>
                            <span className="text-sm font-medium hidden lg:block">Laporan</span>
                        </a>
                        <a className="flex items-center gap-3 px-3 py-3 rounded-xl text-slate-500 dark:text-text-secondary hover:bg-slate-100 dark:hover:bg-card-hover transition-colors" href="/driver-vehicle">
                            <span className="material-symbols-outlined">local_shipping</span>
                            <span className="text-sm font-medium hidden lg:block">Kendaraan</span>
                        </a>
                        <a className="flex items-center gap-3 px-3 py-3 rounded-xl text-slate-500 dark:text-text-secondary hover:bg-slate-100 dark:hover:bg-card-hover transition-colors" href="/user-management">
                            <span className="material-symbols-outlined">group</span>
                            <span className="text-sm font-medium hidden lg:block">Pengguna</span>
                        </a>
                        <a className="flex items-center gap-3 px-3 py-3 rounded-xl text-slate-500 dark:text-text-secondary hover:bg-slate-100 dark:hover:bg-card-hover transition-colors" href="/settings-hardware">
                            <span className="material-symbols-outlined">settings</span>
                            <span className="text-sm font-medium hidden lg:block">Pengaturan</span>
                        </a>
                    </nav>
                    <div className="flex items-center gap-3 p-2 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-card-dark">
                        <div className="bg-center bg-no-repeat bg-cover rounded-full size-10 flex-shrink-0" style={{backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuCekJTzIrsXKhdQS914NutcQOBY6VBllDn79Pxfo4dtR66_dIIh4lmW_xvzXTMmF8EMV_LanWUHsT-uN1HVBiN1mAiGXsAwqZIpAOnajFYM6oOKeekfYwmEPezs6DJ7nkXfbZC9LKkg53HNykWdLADMUo2CEDeV1lUgLHieI1j-o11cNa8wvMfxBL16wTA5OTCAP5s6_zYHLpx8AZJYfR15_AhWYfZksZ33OdrRvx68MMMcr00rjYUIiHSDcTIvFc_gQztTiIh7LFU")'}}></div>
                        <div className="flex flex-col hidden lg:flex overflow-hidden">
                            <p className="text-sm font-bold truncate">Budi Santoso</p>
                            <p className="text-xs text-text-secondary truncate">Operator Pos 1</p>
                        </div>
                    </div>
                </div>
            </aside>
            <main className="flex-1 flex flex-col h-full overflow-hidden bg-background-light dark:bg-[#0b1016]">
                <header className="flex-shrink-0 px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-[#111822]">
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div>
                            <h2 className="text-2xl font-extrabold tracking-tight">Dashboard Pemantauan</h2>
                            <p className="text-text-secondary text-sm mt-1 flex items-center gap-2">
                                <span className="material-symbols-outlined text-base">calendar_today</span>
                                Senin, 23 Oktober 2023 - 10:45 WIB
                            </p>
                        </div>
                        <div className="flex items-center gap-3">
                            <a href="/notifications" className="flex items-center justify-center size-10 rounded-full bg-slate-100 dark:bg-card-hover text-slate-600 dark:text-white hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors relative">
                                <span className="material-symbols-outlined">notifications</span>
                                <span className="absolute top-2 right-2 size-2 bg-red-500 rounded-full"></span>
                            </a>
                            <a href="/weighing-station" className="hidden md:flex items-center gap-2 px-4 py-2 bg-primary hover:bg-blue-600 text-white rounded-lg font-medium text-sm transition-colors">
                                <span className="material-symbols-outlined text-lg">add</span>
                                Input Manual
                            </a>
                        </div>
                    </div>
                </header>
                <div className="flex-1 overflow-y-auto p-4 md:p-6 lg:p-8">
                    <div className="max-w-[1600px] mx-auto flex flex-col gap-6">
                        {/* Stats Cards */}
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 lg:gap-6">
                            <div className="flex flex-col gap-2 rounded-2xl p-6 bg-white dark:bg-[#192433] border border-slate-200 dark:border-slate-800 shadow-sm">
                                <div className="flex items-center justify-between">
                                    <p className="text-text-secondary text-sm font-medium">Total Masuk</p>
                                    <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-500">
                                        <span className="material-symbols-outlined">local_shipping</span>
                                    </div>
                                </div>
                                <div>
                                    <p className="text-3xl font-bold">{stats.total_count} <span className="text-lg font-medium text-text-secondary">Kendaraan</span></p>
                                    <p className="text-emerald-500 text-sm font-medium flex items-center gap-1 mt-1">
                                        <span className="material-symbols-outlined text-base">trending_up</span>
                                        Live Update
                                    </p>
                                </div>
                            </div>
                            <div className="flex flex-col gap-2 rounded-2xl p-6 bg-white dark:bg-[#192433] border border-slate-200 dark:border-slate-800 shadow-sm">
                                <div className="flex items-center justify-between">
                                    <p className="text-text-secondary text-sm font-medium">Total Tonase</p>
                                    <div className="p-2 rounded-lg bg-blue-500/10 text-blue-500">
                                        <span className="material-symbols-outlined">weight</span>
                                    </div>
                                </div>
                                <div>
                                    <p className="text-3xl font-bold">{(stats.total_weight / 1000).toFixed(2)} <span className="text-lg font-medium text-text-secondary">Ton</span></p>
                                    <p className="text-emerald-500 text-sm font-medium flex items-center gap-1 mt-1">
                                        <span className="material-symbols-outlined text-base">trending_up</span>
                                        Accumulated
                                    </p>
                                </div>
                            </div>
                            <div className="flex flex-col gap-2 rounded-2xl p-6 bg-white dark:bg-[#192433] border border-slate-200 dark:border-slate-800 shadow-sm">
                                <div className="flex items-center justify-between">
                                    <p className="text-text-secondary text-sm font-medium">Antrian Saat Ini</p>
                                    <div className="p-2 rounded-lg bg-orange-500/10 text-orange-500">
                                        <span className="material-symbols-outlined">queue</span>
                                    </div>
                                </div>
                                <div>
                                    <p className="text-3xl font-bold">{stats.pending_count} <span className="text-lg font-medium text-text-secondary">Kendaraan</span></p>
                                    <p className="text-text-secondary text-sm font-medium flex items-center gap-1 mt-1">
                                        <span className="material-symbols-outlined text-base">check_circle</span>
                                        Active
                                    </p>
                                </div>
                            </div>
                        </div>

                        {/* Main Content Grid */}
                        <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 h-full">
                            <div className="xl:col-span-8 flex flex-col gap-6">
                                {/* CCTV View */}
                                <div className="rounded-2xl overflow-hidden bg-black relative group shadow-lg">
                                    <div className="absolute top-4 left-4 z-10 px-3 py-1 rounded-full bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-2">
                                        <div className="size-2 rounded-full bg-red-500 animate-pulse"></div>
                                        <span className="text-white text-xs font-medium">CCTV - Gerbang Masuk 1</span>
                                    </div>
                                    <div className="aspect-video w-full bg-cover bg-center" style={{backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuDmQoUSNrPN_R-vwYQXIp1JdhjMT-5vt1jLsa_8Hmc8olvlIw9a7evTJI67enAOU01ECai8ntETUmbzhjpL96BSnuGB2cW5YvI4bKzuxdvTXtPSyYHArhy9qedgtVw0RZUrZT7CMmgqjIfn7sZbqmg6EL6_aDPxsOz4gr7GjNsK__dSyUnP_IOayBmcBIGk8pG2by0ucCbnnhVR7RIQ-ANmADxV2NK7IVNnI-NYn-Qg-XPCj_H187Qrfm7z1VH2VstqV7lQ6Cjpa_I")'}}>
                                        <div className="absolute inset-0 flex items-center justify-center bg-black/20 group-hover:bg-black/10 transition-all">
                                            <button className="flex items-center justify-center rounded-full size-16 bg-white/20 backdrop-blur-sm text-white hover:scale-110 transition-transform">
                                                <span className="material-symbols-outlined text-4xl">play_arrow</span>
                                            </button>
                                        </div>
                                    </div>
                                    <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/90 to-transparent p-6 pt-12">
                                        <div className="flex flex-col md:flex-row items-end md:items-center justify-between gap-4">
                                            <div className="flex items-center gap-4">
                                                <div className="bg-white text-black px-4 py-2 rounded font-mono font-bold text-2xl tracking-wider border-l-8 border-primary shadow-[0_0_15px_rgba(255,255,255,0.3)]">
                                                    B 1234 XYZ
                                                </div>
                                                <div className="flex flex-col">
                                                    <span className="text-green-400 text-xs uppercase font-bold tracking-wider">ANPR Detected</span>
                                                    <span className="text-white/80 text-sm">Confidence: 98.5%</span>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                {/* Weighing Process */}
                                <div className="rounded-2xl bg-white dark:bg-[#192433] border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
                                    <div className="flex items-center justify-between mb-6">
                                        <h3 className="text-lg font-bold flex items-center gap-2">
                                            <span className="material-symbols-outlined text-primary">analytics</span>
                                            Proses Penimbangan
                                        </h3>
                                        <span className="px-3 py-1 rounded-full bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 text-xs font-bold uppercase tracking-wider border border-yellow-500/20">
                                            Stabilizing
                                        </span>
                                    </div>
                                    <div className="flex flex-col md:flex-row gap-8 items-center">
                                        <div className="flex-1 w-full bg-black rounded-xl p-6 border-4 border-slate-700 shadow-inner relative overflow-hidden">
                                            <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
                                            <p className="text-text-secondary text-xs uppercase tracking-widest mb-1 font-mono text-center">Net Weight (KG)</p>
                                            <div className="text-center">
                                                <span className="text-5xl md:text-7xl font-mono font-black text-green-500 tabular-nums tracking-tighter drop-shadow-[0_0_10px_rgba(34,197,94,0.5)]">
                                                    24,500
                                                </span>
                                            </div>
                                            <div className="flex justify-between mt-2 px-4">
                                                <div className="size-2 rounded-full bg-red-900"></div>
                                                <div className="size-2 rounded-full bg-green-500 shadow-[0_0_8px_#22c55e]"></div>
                                                <div className="size-2 rounded-full bg-red-900"></div>
                                            </div>
                                        </div>
                                        <div className="flex-1 w-full flex flex-col gap-4">
                                            <div className="grid grid-cols-2 gap-4">
                                                <div className="p-3 rounded-lg bg-slate-50 dark:bg-[#233348]">
                                                    <p className="text-text-secondary text-xs">Supir</p>
                                                    <p className="font-bold">Ahmad Junaedi</p>
                                                </div>
                                                <div className="p-3 rounded-lg bg-slate-50 dark:bg-[#233348]">
                                                    <p className="text-text-secondary text-xs">Jenis Material</p>
                                                    <p className="font-bold">Batu Split 1/2</p>
                                                </div>
                                                <div className="p-3 rounded-lg bg-slate-50 dark:bg-[#233348]">
                                                    <p className="text-text-secondary text-xs">Vendor</p>
                                                    <p className="font-bold">PT. Sumber Alam</p>
                                                </div>
                                                <div className="p-3 rounded-lg bg-slate-50 dark:bg-[#233348]">
                                                    <p className="text-text-secondary text-xs">PO Number</p>
                                                    <p className="font-bold">#PO-2023-001</p>
                                                </div>
                                            </div>
                                            <div className="flex gap-3 mt-2">
                                                <a href="/weighing-station" className="flex-1 bg-primary hover:bg-blue-600 text-white font-bold py-3 px-4 rounded-xl flex items-center justify-center gap-2 transition-all active:scale-95 shadow-lg shadow-blue-500/20">
                                                    <span className="material-symbols-outlined">save</span>
                                                    Simpan Data
                                                </a>
                                                <button className="bg-slate-100 dark:bg-[#233348] hover:bg-slate-200 dark:hover:bg-[#2d4059] text-slate-700 dark:text-white font-bold py-3 px-4 rounded-xl flex items-center justify-center gap-2 transition-colors">
                                                    <span className="material-symbols-outlined">print</span>
                                                </button>
                                                <button className="bg-slate-100 dark:bg-[#233348] hover:bg-slate-200 dark:hover:bg-[#2d4059] text-slate-700 dark:text-white font-bold py-3 px-4 rounded-xl flex items-center justify-center gap-2 transition-colors text-red-500">
                                                    <span className="material-symbols-outlined">restart_alt</span>
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className="xl:col-span-4 flex flex-col gap-6 h-full">
                                <div className="flex items-center justify-between">
                                    <h3 className="text-lg font-bold">Antrian & Riwayat</h3>
                                    <a href="/history-detail" className="text-primary text-sm font-medium hover:underline">Lihat Semua</a>
                                </div>
                                <div className="flex flex-col gap-3 flex-1 overflow-hidden">
                                    {/* Dynamic List */}
                                    {recentTx.length === 0 ? (
                                        <p className="text-center text-slate-500 p-4">Belum ada data</p>
                                    ) : recentTx.map((tx, idx) => (
                                        <div key={tx.ticket_id} className="p-4 rounded-xl bg-white dark:bg-[#151e29] border border-transparent hover:bg-slate-50 dark:hover:bg-[#1b2633] transition-colors group">
                                            <div className="flex justify-between items-center">
                                                <div className="flex items-center gap-3">
                                                    <div className="size-8 rounded-full bg-emerald-500/20 text-emerald-500 flex items-center justify-center">
                                                        <span className="material-symbols-outlined text-lg">{tx.status === 'PENDING' ? 'pending' : 'check'}</span>
                                                    </div>
                                                    <div>
                                                        <h5 className="text-sm font-bold text-slate-800 dark:text-slate-200">{tx.plate_number}</h5>
                                                        <p className="text-xs text-text-secondary">{(tx.net_weight/1000).toFixed(2)} Ton</p>
                                                    </div>
                                                </div>
                                                <div className="text-right">
                                                    <span className={`px-2 py-1 rounded text-[10px] font-bold ${tx.status === 'PENDING' ? 'bg-yellow-500/10 text-yellow-500' : 'bg-emerald-500/10 text-emerald-500'}`}>{tx.status}</span>
                                                    <p className="text-xs text-text-secondary mt-1">{new Date(tx.entry_time).toLocaleTimeString()}</p>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};

// Screen: Weighing Station
const WeighingStationScreen = () => {
    return (
        <div className="bg-background-light dark:bg-background-dark text-slate-900 dark:text-white font-display overflow-x-hidden min-h-screen flex flex-col">
            <header className="sticky top-0 z-50 flex items-center justify-between border-b border-solid border-slate-200 dark:border-border-dark bg-white dark:bg-[#151f2e] px-6 py-3 shadow-sm">
                <div className="flex items-center gap-4">
                    <div className="flex items-center justify-center size-10 rounded-lg bg-primary/10 text-primary">
                        <span className="material-symbols-outlined">local_shipping</span>
                    </div>
                    <div>
                        <h2 className="text-lg font-bold leading-tight tracking-tight">Stasiun Penimbangan 01</h2>
                        <div className="flex items-center gap-2 text-xs font-medium text-slate-500 dark:text-slate-400">
                            <span className="flex items-center gap-1"><span className="block w-2 h-2 rounded-full bg-green-500"></span> Online</span>
                            <span className="text-slate-300 dark:text-slate-600">|</span>
                            <span>Operator: Budi Santoso</span>
                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-6">
                    <div className="hidden md:flex gap-4 text-xs font-semibold">
                        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-100 dark:bg-surface-dark border border-slate-200 dark:border-border-dark">
                            <span className="material-symbols-outlined text-green-500 text-sm">videocam</span>
                            <span>CCTV: OK</span>
                        </div>
                        {/* More indicators */}
                    </div>
                    <div className="h-8 w-[1px] bg-slate-200 dark:bg-border-dark hidden md:block"></div>
                    <div className="flex gap-2">
                         <a href="/settings-hardware" className="flex items-center justify-center size-10 rounded-lg bg-slate-100 dark:bg-surface-dark text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-border-dark transition-colors">
                            <span className="material-symbols-outlined">settings</span>
                        </a>
                        <a href="/" className="flex items-center justify-center size-10 rounded-lg bg-slate-100 dark:bg-surface-dark text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-border-dark transition-colors">
                            <span className="material-symbols-outlined">logout</span>
                        </a>
                    </div>
                </div>
            </header>
            <main className="flex-1 p-4 md:p-6 lg:p-8 max-w-[1920px] mx-auto w-full">
                <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 h-full">
                    <div className="xl:col-span-7 flex flex-col gap-6">
                         {/* Live Feed */}
                         <div className="flex flex-col gap-4 bg-white dark:bg-surface-dark p-4 rounded-xl border border-slate-200 dark:border-border-dark shadow-sm">
                             {/* ... Video Content ... */}
                             <div className="relative w-full aspect-video bg-black rounded-lg overflow-hidden group">
                                <div className="absolute inset-0 bg-cover bg-center opacity-80" style={{backgroundImage: "url('https://lh3.googleusercontent.com/aida-public/AB6AXuBps0zUekUOB9LpL2aAhj3ZwNjUrKxlpk7oUA5Y5s3KW6wQLo0eG3IADp0AX5tFyAr10ntUyZ78NZl8FgEcDOHqDd44ZELw2L79i87NJ5gHZQQIZhf9piyJNjM8stX1KAQxePQHrCa_X1e5E4o75qbrzUgExwEsgwNC1hL2oAZ9kD-sd6mvRitXrW7TBp-3Q6DWXnv_x0ZHUGSxxLXTtU5LtCcFUXJWDk6kEm5GJHyvrDqrGNBOJaoF5aXcWRU6isPakyE30EPzfLU')"}}></div>
                             </div>
                             {/* ... */}
                         </div>
                    </div>
                    <div className="xl:col-span-5 flex flex-col gap-6">
                        <div className="bg-[#111822] rounded-2xl p-6 border border-border-dark shadow-lg relative overflow-hidden">
                             {/* Weight Display */}
                             <div className="relative z-10 py-4 text-center">
                                <h1 className="text-6xl md:text-7xl font-black text-white tracking-tighter tabular-nums leading-none">
                                    24,580
                                </h1>
                                <p className="text-primary font-bold text-xl mt-2">KILOGRAM (KG)</p>
                            </div>
                        </div>
                        {/* Form */}
                         <div className="bg-white dark:bg-surface-dark p-6 rounded-xl border border-slate-200 dark:border-border-dark shadow-sm flex-1 flex flex-col">
                            <h3 className="text-lg font-bold mb-6 flex items-center gap-2 pb-4 border-b border-slate-200 dark:border-border-dark">
                                <span className="material-symbols-outlined text-primary">edit_document</span>
                                Input Transaksi
                            </h3>
                            <div className="flex flex-col gap-5 flex-1">
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                                        Nomor Polisi <span className="text-primary text-xs ml-1">(Otomatis)</span>
                                    </label>
                                    <input className="block w-full rounded-lg border-slate-300 dark:border-border-dark bg-slate-50 dark:bg-[#111822] text-slate-900 dark:text-white shadow-sm focus:border-primary focus:ring-primary sm:text-lg font-bold p-3" readOnly type="text" value="B 9821 UI"/>
                                </div>
                                <button
                                    onClick={() => {
                                        fetch('/api/transactions', {
                                            method: 'POST',
                                            headers: {'Content-Type': 'application/json'},
                                            body: JSON.stringify({
                                                plate_number: "B 9821 UI",
                                                driver_name: "Ahmad Junaedi", // Mocked from UI
                                                vendor: "PT. Sumber Alam",
                                                net_weight: 24580
                                            })
                                        })
                                        .then(res => res.json())
                                        .then(data => {
                                            alert("Data saved! Ticket ID: " + data.ticket_id);
                                            if (data.pdf_url) {
                                                window.open(data.pdf_url, '_blank');
                                            }
                                        })
                                        .catch(err => alert("Error saving data"));
                                    }}
                                    className="w-full bg-primary hover:bg-blue-600 text-white font-bold py-4 px-6 rounded-lg shadow-lg shadow-primary/20 flex items-center justify-center gap-3 transition-transform active:scale-[0.98]">
                                    <span className="material-symbols-outlined">save</span>
                                    SIMPAN BERAT (F2)
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};
