# MW Toolkit - Go & MySQL Task Management Web Application

MW Toolkit adalah aplikasi manajemen tugas berbasis web yang dibuat menggunakan bahasa pemrograman **Go (Golang)** dan database **MySQL**. Aplikasi ini didesain dengan antarmuka pengguna modern (*dark mode*) yang responsif dan interaktif, lengkap dengan grafik visual statistik untuk membantu memantau produktivitas Anda.

## Fitur Utama
1. **Autentikasi Pengguna**: Sistem pendaftaran akun (Register), masuk (Login) aman menggunakan hashing kata sandi `bcrypt`, dan keluar (Logout) dengan session berbasis cookie.
2. **Manajemen Tugas (CRUD)**:
   - Menambahkan tugas dengan judul, deskripsi, kategori (Work, Personal, Urgent, dsb), sumber tugas (Manual, Email, Slack, WhatsApp, dsb), status, dan tenggat waktu (*due date*).
   - Memperbarui tugas secara dinamis menggunakan pop-up modal tanpa memuat ulang halaman (*full page reload*).
   - Menghapus tugas dengan konfirmasi keamanan.
   - Deteksi otomatis tugas yang melewati tenggat waktu (*overdue*).
3. **Pencarian & Penyaringan**: Menyaring tugas berdasarkan status, kategori, atau mencari kata kunci pada judul dan deskripsi secara instan.
4. **Dashboard Statistik (Grafik)**:
   - **KPI Summary**: Menampilkan total tugas, tugas tertunda, sedang berjalan, selesai, serta rasio penyelesaian dalam bentuk progress bar.
   - **Tugas by Kategori**: Grafik Doughnut interaktif yang menampilkan proporsi kategori tugas.
   - **Tugas by Status**: Grafik Bar yang menunjukkan sebaran status tugas.
   - **Tren Pembuatan Tugas**: Grafik Line dengan efek gradien glowing yang menunjukkan aktivitas penambahan tugas selama 14 hari terakhir.

---

## Prasyarat
- **Go (Golang)**: Versi 1.22 atau yang lebih baru.
- **MySQL**: Server MySQL yang aktif (misal: MySQL Installer, XAMPP, Laragon, dsb).

---

## Panduan Instalasi & Konfigurasi

### 1. Konfigurasi Environment (`.env`)
Di dalam direktori proyek, terdapat file `.env`. Anda perlu menyesuaikan kredensial MySQL lokal Anda di dalam file tersebut:
```env
PORT=8080
DB_USER=root            # Username MySQL Anda
DB_PASSWORD=yourpassword # Password MySQL Anda (kosongkan jika tidak ada password)
DB_NAME=task_manager    # Nama database (akan dibuat otomatis oleh program)
DB_HOST=127.0.0.1
DB_PORT=3306
SESSION_SECRET=a-very-secure-32-byte-long-session-secret-key
```

*Catatan: Program akan otomatis mendeteksi jika database `task_manager` belum ada dan akan membuatnya beserta seluruh tabel yang diperlukan pada saat program dijalankan.*

### 2. Menjalankan Aplikasi
Buka terminal/PowerShell di direktori proyek `task-manager-go` dan jalankan perintah:
```bash
go run .
```

Jika berhasil, Anda akan melihat log seperti berikut:
```text
2026/05/24 19:58:00 Configuration loaded successfully
2026/05/24 19:58:01 Database 'task_manager' verified/created successfully
2026/05/24 19:58:01 Database connection pool established successfully
2026/05/24 19:58:01 Table 'users' verified/created
2026/05/24 19:58:01 Table 'tasks' verified/created
2026/05/24 19:58:01 Server starting on http://localhost:8080
```

### 3. Akses Aplikasi
Buka browser Anda dan akses:
[http://localhost:8080](http://localhost:8080)

---

## Struktur Direktori Proyek
- `main.go`: Titik masuk aplikasi dan registrasi rute HTTP.
- `config/`: Memuat konfigurasi variabel lingkungan (.env).
- `db/`: Pengaturan koneksi MySQL dan auto-migrasi skema database.
- `models/`: Struktur data dan kueri database (User & Task).
- `handlers/`: Logika pengontrol alur (Auth, Task, Dashboard, Render).
- `middleware/`: Middleware otentikasi sesi rute.
- `templates/`: Template HTML (layout, login, register, dashboard, tasks).
- `static/`: Aset statis seperti CSS (gaya tampilan premium) dan JS (interaksi modal).
