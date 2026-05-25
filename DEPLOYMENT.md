# Panduan Panduan Deployment Ke Google Cloud Platform (GCP)

Dokumen ini menjelaskan dua metode untuk menjalankan aplikasi **TaskFlow** secara online di Google Cloud Platform (GCP). Karena tools `gcloud` dan `docker` tidak terinstal secara lokal, metode ini dirancang untuk dijalankan sepenuhnya menggunakan **Google Cloud Shell** (terminal web gratis yang disediakan oleh Google Cloud Console yang sudah terinstal `gcloud`, `docker`, dan `git`).

---

## Persiapan Awal (Google Cloud Shell)

1. Buka [GCP Console](https://console.cloud.google.com/).
2. Buat proyek baru atau pilih proyek yang sudah ada. Catat **Project ID** Anda.
3. Klik tombol **Activate Cloud Shell** di pojok kanan atas layar (ikon `>_`).
4. Clone repositori atau unggah file aplikasi ke Cloud Shell Anda.

---

## Opsi A: Menggunakan Google Cloud Run + Cloud SQL (Serverless & Skala Otomatis)

Opsi ini sangat cocok untuk produksi. Cloud Run bersifat serverless dan gratis untuk trafik rendah, namun database Cloud SQL akan mengenakan biaya bulanan kecil (mulai dari ~$10/bulan).

### Langkah 1: Aktifkan API yang Dibutuhkan
Jalankan perintah berikut di Cloud Shell:
```bash
gcloud services enable run.googleapis.com \
                       sqladmin.googleapis.com \
                       cloudbuild.googleapis.com
```

### Langkah 2: Buat Instance Cloud SQL (MySQL)
Buat instance database MySQL mikro baru:
```bash
gcloud sql instances create taskflow-db \
    --database-version=MYSQL_8_0 \
    --tier=db-f1-micro \
    --region=us-central1
```

Setelah pembuatan selesai, buat database dan atur password root:
```bash
# Ganti SECURE_PASSWORD dengan password pilihan Anda
gcloud sql users set-password root --instance=taskflow-db --password=SECURE_PASSWORD

# Buat database bernama task_manager
gcloud sql databases create task_manager --instance=taskflow-db
```

Dapatkan nama koneksi instance Anda:
```bash
gcloud sql instances describe taskflow-db --format="value(connectionName)"
# Output akan berformat: PROJECT_ID:REGION:taskflow-db
```

### Langkah 3: Build Container Menggunakan Cloud Build
Build image container Docker langsung di cloud tanpa menginstal Docker secara lokal:
```bash
gcloud builds submit --tag gcr.io/$GOOGLE_CLOUD_PROJECT/taskflow:v1
```

### Langkah 4: Deploy ke Google Cloud Run
Deploy aplikasi dan hubungkan ke Cloud SQL instance. 
> [!IMPORTANT]
> Masukkan nama koneksi instance dari Langkah 2 ke parameter `--add-cloudsql-instances` dan atur `DB_HOST` sebagai `/cloudsql/PROJECT_ID:REGION:taskflow-db` (dengan prefix `/cloudsql/` untuk koneksi Unix Socket yang aman).

```bash
# Ganti PROJECT_ID:REGION:taskflow-db dan SECURE_PASSWORD
gcloud run deploy taskflow \
    --image gcr.io/$GOOGLE_CLOUD_PROJECT/taskflow:v1 \
    --add-cloudsql-instances=PROJECT_ID:REGION:taskflow-db \
    --update-env-vars=DB_USER=root,DB_PASSWORD=SECURE_PASSWORD,DB_NAME=task_manager,DB_HOST=/cloudsql/PROJECT_ID:REGION:taskflow-db,SESSION_SECRET=ganti-dengan-kunci-rahasia-anda \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated
```

Setelah selesai, Cloud Run akan memberikan URL publik untuk aplikasi Anda (misal: `https://taskflow-xxxxxx-uc.a.run.app`).

---

## Opsi B: Menggunakan Compute Engine Always-Free VM (100% Gratis Selamanya)

Google Cloud menawarkan VM gratis selamanya (`e2-micro` dengan memori 1GB) di region AS tertentu (Oregon `us-west1`, Iowa `us-central1`, South Carolina `us-east1`). Anda dapat menjalankan server Go dan database MySQL secara bersamaan pada VM ini.

### Langkah 1: Buat Instance VM Gratis
Jalankan perintah ini di Cloud Shell:
```bash
gcloud compute instances create taskflow-free-vm \
    --machine-type=e2-micro \
    --zone=us-central1-a \
    --image-family=debian-12 \
    --image-project=debian-cloud \
    --boot-disk-size=10GB \
    --tags=http-server
```

### Langkah 2: Buat Aturan Firewall untuk Port 8080
Aplikasi kita berjalan pada port `8080`. Buat aturan firewall untuk mengizinkan akses publik ke port tersebut:
```bash
gcloud compute firewall-rules create allow-http-8080 \
    --direction=INGRESS \
    --priority=1000 \
    --network=default \
    --action=ALLOW \
    --rules=tcp:8080 \
    --source-ranges=0.0.0.0/0 \
    --target-tags=http-server
```

### Langkah 3: Hubungkan ke VM dan Instal Environment
Hubungkan ke VM melalui SSH:
```bash
gcloud compute ssh taskflow-free-vm --zone=us-central1-a
```

Setelah masuk ke dalam shell VM, instal Go, MySQL, dan Git:
```bash
sudo apt update
sudo apt install -y golang-go mariadb-server git
```

### Langkah 4: Konfigurasi Database di VM
Mulai layanan database MariaDB/MySQL dan konfigurasikan:
```bash
sudo systemctl start mariadb
sudo systemctl enable mariadb

# Masuk ke MySQL shell untuk membuat database
sudo mysql -u root -e "CREATE DATABASE task_manager;"
sudo mysql -u root -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '07Mei2015';"
sudo mysql -u root -e "FLUSH PRIVILEGES;"
```

### Langkah 5: Clone dan Jalankan Aplikasi di VM
Unduh aplikasi dan jalankan:
```bash
git clone <URL_REPOSITORI_ANDA> taskflow
cd taskflow

# Buat file konfigurasi .env
cat <<EOT > .env
PORT=8080
DB_USER=root
DB_PASSWORD=07Mei2015
DB_NAME=task_manager
DB_HOST=127.0.0.1
DB_PORT=3306
SESSION_SECRET=kunci-rahasia-sangat-aman-12345
EOT

# Build dan jalankan aplikasi di latar belakang
go build -o taskflow-app .
nohup ./taskflow-app > app.log 2>&1 &
```

Dapatkan IP Publik VM Anda:
```bash
# Jalankan perintah ini di luar VM (atau di Cloud Shell)
gcloud compute instances describe taskflow-free-vm \
    --zone=us-central1-a \
    --format="value(networkInterfaces[0].accessConfigs[0].natIP)"
```
Buka `http://<IP_PUBLIK_VM>:8080` pada browser Anda untuk mengakses aplikasi.
