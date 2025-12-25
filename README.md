# StoneWeigh - System Timbangan Digital Modern

StoneWeigh adalah aplikasi manajemen jembatan timbang (weighbridge) modern yang dibangun menggunakan **Go (Golang)**, **HTML Templates**, dan **Tailwind CSS**. System ini dirancang untuk kecepatan, keandalan offline, dan kemudahan integrasi hardware.

## 🚀 Fitur Utama

*   **Tampilan Modern & Responsif**: Menggunakan Tailwind CSS dengan dukungan Dark Mode.
*   **Integrasi Hardware**:
    *   Mendukung hingga 3 indikator timbangan via Serial (RS232).
    *   Integrasi CCTV untuk deteksi plat nomor otomatis (ANPR) menggunakan OpenCV.
*   **Manajemen Data**:
    *   Dashboard analitik harian (Total kendaraan, total berat).
    *   Master Data Kendaraan & Supir.
    *   Pencetakan tiket/surat jalan dalam format PDF.
*   **Keamanan**:
    *   Login Administrator & Operator.
    *   Rate Limiting & Logging.
    *   Data tersimpan aman di database (SQLite/PostgreSQL).

## 🛠️ Teknologi

*   **Backend**: Go 1.22+ (Gin Framework, GORM)
*   **Frontend**: Server-side HTML Templates + Tailwind CSS (via CDN/Local)
*   **Database**: SQLite (Default) atau PostgreSQL
*   **Computer Vision**: GoCV (OpenCV bindings) atau Mock Mode
*   **PDF**: gofpdf

## 📦 Instalasi

### Prasyarat
1.  **Go** (Versi 1.21 atau lebih baru)
2.  **(Opsional)** **OpenCV 4.x** jika ingin menggunakan fitur ANPR asli. Jika tidak, aplikasi akan berjalan dalam mode Mock (Simulasi).

### Langkah-langkah

1.  **Clone Repository**
    ```bash
    git clone https://github.com/username/stoneweigh.git
    cd stoneweigh
    ```

2.  **Konfigurasi Environment**
    Salin file contoh konfigurasi dan sesuaikan:
    ```bash
    cp .env.example .env
    ```
    Edit `.env` untuk mengatur port serial timbangan, database, dan akun admin awal.

3.  **Jalankan Aplikasi**

    *   **Mode Standard (Tanpa OpenCV/ANPR)**:
        Gunakan tag build default (Mock ANPR akan aktif jika library tidak ditemukan, atau paksa dengan tag custom jika diimplementasikan). Saat ini, aplikasi akan otomatis menggunakan Mock jika build tag `gocv` tidak disertakan atau jika library system tidak ada (tergantung implements file).

        Untuk lingkungan pengembangan tanpa OpenCV:
        ```bash
        go run cmd/server/main.go
        ```

    *   **Mode Produksi (Dengan OpenCV/ANPR)**:
        Pastikan OpenCV terinstall, lalu jalankan:
        ```bash
        go run -tags gocv cmd/server/main.go
        ```

4.  **Akses Aplikasi**
    Buka browser dan kunjungi: `http://localhost:8080`

## ⚙️ Konfigurasi Hardware

### Timbangan (Serial / RS232)
Edit `.env` untuk menentukan port COM/TTY:
```ini
SCALE_PORTS="1=COM3:9600,2=/dev/ttyUSB0:9600"
```
Jika tidak ada timbangan fisik, Anda dapat mengaktifkan **Demo Mode** untuk simulasi berat server-side:
```bash
export ENABLE_DEMO_SCALE=true
```

### CCTV / ANPR
Edit `.env` untuk URL RTSP kamera:
```ini
CCTV_RTSP_URLS="1=rtsp://admin:pass@192.168.1.10:554/stream"
```
Model deteksi plat nomor (`.pt`) harus diletakkan di folder `models/platdetection.pt`.

## 📚 Panduan Penggunaan

1.  **Login**: Masuk menggunakan kredensial admin (Default: `admin` / `secret_password_123` dari .env).
2.  **Dashboard**: Lihat ringkasan transaksi hari ini.
3.  **Pengaturan**: Daftarkan kendaraan dan supir rutin di menu "Manajemen Kendaraan" untuk mempercepat proses input.
4.  **Timbangan**:
    *   Pilih timbangan yang aktif.
    *   Berat akan muncul secara real-time.
    *   Klik "Photo & Analisa" untuk menangkap plat nomor.
    *   Isi detail muatan dan klik "Simpan".
5.  **Laporan**: Buka menu Laporan untuk melihat riwayat dan mencetak ulang PDF.

## 📁 Struktur Project

```
stoneweigh/
├── cmd/server/         # Entry point aplikasi
├── internal/
│   ├── cv/             # Logika Computer Vision (ANPR)
│   ├── handlers/       # HTTP Handlers (Controller)
│   ├── hardware/       # Driver Serial Timbangan
│   ├── middleware/     # Auth, Logging, RateLimit
│   ├── models/         # Database Structs
│   ├── reporting/      # Generator PDF
│   └── router/         # Konfigurasi Gin Router
├── models/             # File model AI (.pt)
├── web/
│   ├── static/         # CSS, JS, Images
│   └── templates/      # File HTML
└── .env.example        # Template konfigurasi
```

## 📄 Lisensi
[MIT License](LICENSE)
