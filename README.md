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
    *   **Manajemen Pengguna & Stasiun**: Atur siapa yang mengelola timbangan tertentu.
    *   **Konfigurasi Dinamis**: Ubah setting hardware tanpa restart via UI.
    *   Pencetakan tiket/surat jalan dalam format PDF.
*   **Keamanan**:
    *   Login Administrator & Operator.
    *   System Logs Viewer.
    *   Rate Limiting & Logging.
    *   Data tersimpan aman di PostgreSQL (Disarankan) atau SQLite.

## 🛠️ Teknologi

*   **Backend**: Go 1.22+ (Gin Framework, GORM)
*   **Frontend**: Server-side HTML Templates + Tailwind CSS (via CDN/Local)
*   **Database**: SQLite (Default) atau PostgreSQL
*   **Computer Vision**: GoCV (OpenCV bindings) atau Mock Mode
*   **PDF**: gofpdf

## 📦 Instalasi OpenCV (Production)

Untuk menggunakan fitur deteksi plat nomor (ANPR) di lingkungan produksi, Anda wajib menginstall OpenCV 4.x.

### 🐧 Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install -y libopencv-dev build-essential
```

### 🪟 Windows
1. Download dan Install [MinGW-W64](https://sourceforge.net/projects/mingw-w64/).
2. Download OpenCV Prebuilt for Windows atau gunakan Chocolatey:
   ```powershell
   choco install opencv
   ```
3. Pastikan `OPENCV_DIR` terdaftar di Environment Variables.

### 🍎 macOS
```bash
brew install opencv
```

---

## 🚀 Cara Menjalankan

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
    **PENTING**: Edit `.env` untuk mengatur koneksi **PostgreSQL**. Jika gagal connect ke Postgres, aplikasi akan berhenti (safety measure).

3.  **Jalankan Aplikasi**

    *   **Mode Standard (Mock ANPR)**:
        Jika OpenCV tidak terinstall, ANPR akan menggunakan mode simulasi.
        ```bash
        go run cmd/server/main.go
        ```

    *   **Mode Produksi (Dengan OpenCV)**:
        Pastikan OpenCV terinstall dengan benar.
        ```bash
        go run -tags gocv cmd/server/main.go
        ```

4.  **Akses Aplikasi**
    Buka browser dan kunjungi: `http://localhost:8080`

## ⚙️ Panduan Konfigurasi

### 1. Database
Pastikan PostgreSQL sudah berjalan. Buat database baru (misal: `stone`).
Setting di `.env`:
```ini
DB_DRIVER=postgres
DB_DSN="host=localhost user=postgres password=secret dbname=stone port=5432 sslmode=disable"
```

### 2. Hardware (Timbangan & CCTV)
Sekarang Anda dapat mengatur hardware langsung dari aplikasi!
1. Login sebagai **Admin**.
2. Masuk ke **Pengaturan > Konfigurasi Hardware**.
3. Tambah Stasiun baru, masukkan Port Serial (contoh: `/dev/ttyUSB0` atau `COM3`) dan URL RTSP CCTV.

### 3. Manajemen User & Akses
Anda dapat membatasi operator hanya bisa mengakses timbangan tertentu.
1. Masuk ke **Pengaturan > Manajemen Pengguna**.
2. Buat User baru (Role: Operator).
3. Klik **"Atur Akses"** dan pilih timbangan yang diizinkan untuk user tersebut.

### 4. System Logs
Untuk memantau error atau aktivitas sistem:
1. Masuk ke **Pengaturan > Log Sistem**.
2. Log akan refresh otomatis setiap 5 detik.

## 📁 Struktur Project

```
stoneweigh/
├── cmd/server/         # Entry point aplikasi
├── internal/
│   ├── cv/             # Logika Computer Vision (ANPR)
│   ├── handlers/       # HTTP Handlers (Controller)
│   ├── hardware/       # Driver Serial Timbangan
│   ├── pkg/logger/     # System Logger
│   ├── models/         # Database Structs
│   ├── reporting/      # Generator PDF
│   └── router/         # Konfigurasi Gin Router
├── web/
│   ├── static/         # CSS, JS, Images, Reports
│   └── templates/      # File HTML
└── .env.example        # Template konfigurasi
```

## 📄 Lisensi
[MIT License](LICENSE)
