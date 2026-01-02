# Panduan Deployment StoneWeigh ke VPS

Dokumen ini menjelaskan langkah-langkah untuk mendeploy aplikasi StoneWeigh ke VPS (Virtual Private Server) dan menghubungkannya dengan domain menggunakan Nginx dan HTTPS (SSL).

## Prasyarat

1.  **VPS** dengan sistem operasi Linux (Ubuntu 20.04/22.04 LTS atau Debian 11/12 direkomendasikan).
2.  **Domain** yang sudah diarahkan (A Record) ke IP Address VPS Anda.
3.  Akses **SSH** root atau user dengan hak sudo ke VPS.

## Langkah 1: Persiapan Lingkungan VPS

Masuk ke VPS Anda via SSH dan update sistem serta install Docker & Docker Compose.

```bash
# Update repository dan paket sistem
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose (biasanya sudah include di versi docker terbaru sebagai plugin)
sudo apt install docker-compose-plugin -y

# Verifikasi instalasi
docker compose version
```

## Langkah 2: Clone Repository & Konfigurasi

Clone source code aplikasi ke direktori yang diinginkan (misal: `/opt/stoneweigh` atau `~/stoneweigh`).

```bash
# Clone repository (ganti URL dengan repository Anda)
git clone https://github.com/username/stoneweigh.git
cd stoneweigh

# Buat file konfigurasi .env dari template atau manual
cp .env.example .env  # Jika ada, atau buat baru
nano .env
```

**Konfigurasi Penting di `.env`:**
Pastikan Anda mengubah nilai default untuk keamanan produksi:

```ini
# Gunakan mode release untuk produksi
GIN_MODE=release
PORT=8080

# Security
SESSION_SECRET=GANTI_DENGAN_STRING_ACAK_YANG_PANJANG_DAN_RUMIT_!@#

# Admin Awal
ADMIN_USERNAME=admin
ADMIN_PASSWORD=password_yang_sangat_kuat

# Database (Default SQLite sudah cukup untuk skala kecil/menengah)
DB_DRIVER=sqlite
DB_DSN=stoneweigh.db
```

## Langkah 3: Menjalankan Aplikasi dengan Docker

Jalankan aplikasi menggunakan Docker Compose. Proses ini akan membuild image dan menjalankan container.

```bash
# Build dan jalankan di background (detached mode)
docker compose up -d --build

# Cek status container
docker compose ps

# Cek logs jika ada error
docker compose logs -f
```

Saat ini aplikasi sudah berjalan di `http://IP-VPS:8080`.

## Langkah 4: Setup Nginx sebagai Reverse Proxy

Untuk menghubungkan domain dan menghilangkan port 8080 di URL, kita gunakan Nginx.

```bash
# Install Nginx
sudo apt install nginx -y
```

Buat konfigurasi server block baru untuk domain Anda:

```bash
sudo nano /etc/nginx/sites-available/stoneweigh
```

Isi dengan konfigurasi berikut (ganti `domain-anda.com` dengan domain asli):

```nginx
server {
    listen 80;
    server_name domain-anda.com www.domain-anda.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Optional: Cache control untuk static files
    location /static/ {
        proxy_pass http://localhost:8080;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }
}
```

Aktifkan konfigurasi dan restart Nginx:

```bash
# Link ke sites-enabled
sudo ln -s /etc/nginx/sites-available/stoneweigh /etc/nginx/sites-enabled/

# Cek syntax error
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

Sekarang akses `http://domain-anda.com`. Seharusnya aplikasi sudah muncul.

## Langkah 5: Pasang SSL (HTTPS) dengan Certbot

Amankan koneksi menggunakan SSL gratis dari Let's Encrypt.

```bash
# Install Certbot dan plugin Nginx
sudo apt install certbot python3-certbot-nginx -y

# Request sertifikat (ikuti instruksi di layar)
sudo certbot --nginx -d domain-anda.com -d www.domain-anda.com
```

Certbot akan otomatis memodifikasi file Nginx Anda untuk mengalihkan HTTP ke HTTPS.

## Langkah 6: Maintenance & Update

**Melihat Logs:**
```bash
cd ~/stoneweigh
docker compose logs -f --tail=100
```

**Update Aplikasi:**
Jika ada perubahan kode di repository:
```bash
cd ~/stoneweigh
git pull origin main
docker compose up -d --build
docker image prune -f  # Hapus image lama yang tidak terpakai
```

**Backup Database (SQLite):**
Cukup copy file `data/stoneweigh.db` (jika dimount di volume) atau file di dalam container.
Sesuai `docker-compose.yml`, volume `./data` dimount, jadi database aman di host.

```bash
# Contoh backup manual
cp data/stoneweigh.db data/backup_stoneweigh_$(date +%F).db
```

## Troubleshooting Umum

1.  **Camera RTSP tidak muncul:**
    *   Pastikan VPS memiliki koneksi ke kamera jika kamera berada di jaringan publik, atau gunakan VPN jika kamera di jaringan lokal (Site-to-Site VPN).
    *   Cek logs: `docker compose logs | grep gst` untuk melihat error GStreamer.

2.  **Error 502 Bad Gateway:**
    *   Artinya Nginx tidak bisa menghubungi aplikasi Docker. Cek apakah container mati: `docker compose ps`.

3.  **Permission Error pada Logs:**
    *   Pastikan folder logs di host bisa ditulisi: `chmod 777 logs`.
