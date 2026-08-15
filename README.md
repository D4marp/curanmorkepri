# CURANMOR AI — Database Curanmor Online Ditreskrimum Polda Kepulauan Riau

Sistem basis data & dashboard web untuk pencatatan dan pencarian data kendaraan
bermotor curian (curanmor), terintegrasi dengan API untuk chatbot WhatsApp AI.
Dibangun sesuai skema pada *"Susunan Database CURANMOR AI"* dan diskusi
kebutuhan (multi-tenant Polda/Polres/Polsek, pencarian rentang waktu, peta
sebaran, RBAC 6 peran).

**Stack**: Go 1.24 (stdlib `net/http`, tanpa framework berat) + PostgreSQL 16 +
HTML/CSS/JS vanilla (Leaflet untuk peta, Chart.js untuk grafik — keduanya
di-*vendor* lokal, tanpa CDN eksternal). Satu image Docker, satu
`docker compose up` untuk menjalankan semuanya.

---

## 1. Cara Menjalankan (Docker Compose)

```bash
cp .env.example .env
# Edit .env — WAJIB isi POSTGRES_PASSWORD, JWT_SECRET, APP_ENCRYPTION_KEY
# Cara cepat membuat nilai acak yang kuat:
openssl rand -hex 32   # jalankan 2x, satu untuk JWT_SECRET, satu untuk APP_ENCRYPTION_KEY
openssl rand -base64 24  # untuk POSTGRES_PASSWORD

docker compose up -d --build
```

Setelah kontainer `api` sehat (healthcheck OK), buka `http://localhost:8080`
(atau port sesuai `HTTP_PUBLISH_PORT`). Migrasi skema database & seed data
awal (satuan kerja, peran, RBAC, akun Super Admin) berjalan **otomatis** saat
kontainer `api` pertama kali start.

> **Catatan HTTPS**: cookie sesi login memakai flag `Secure` secara default
> (`COOKIE_SECURE=true`), artinya browser hanya mau mengirim cookie ini lewat
> HTTPS. Untuk uji coba lokal tanpa TLS, set `COOKIE_SECURE=false` di `.env`.
> Untuk produksi, **wajib** pasang reverse proxy (nginx/Caddy) dengan
> sertifikat TLS di depan container `api`, lalu biarkan `COOKIE_SECURE=true`.

### Login pertama kali

| NRP | Password |
|---|---|
| `00000000` | `ChangeMe!12345` |

**Wajib langsung diganti** setelah login pertama (menu profil → Ganti Sandi),
dan nomor WhatsApp dummy `0812xxxx1008` pada whitelist juga wajib
diperbarui/dihapus sebelum sistem dipakai operasional (lihat
`scripts/02_seed_rbac.sql`).

---

## 2. Struktur Folder

```
curanmor-ai/
├── cmd/api/main.go              Entry point, wiring seluruh route & middleware
├── internal/
│   ├── config/                  Loader environment variable
│   ├── db/                      Koneksi Postgres + migration/seed runner
│   ├── auth/                    JWT (HS256, implementasi sendiri, tanpa dependensi),
│   │                            bcrypt password hashing, API key generation
│   ├── cryptox/                 Enkripsi AES-256-GCM (dipakai untuk NIK at-rest)
│   ├── middleware/               AuthN (cookie/Bearer JWT), API key (WA), RBAC per
│   │                            modul, tenant scoping, audit trail, rate limit, CORS,
│   │                            security headers
│   ├── httpx/                   Router tipis di atas net/http.ServeMux (Go 1.22+),
│   │                            response envelope standar
│   ├── models/                  Struct data per entitas (20 entitas, 4 domain)
│   ├── repo/                    Akses data (SQL murni, database/sql + lib/pq)
│   └── handlers/                HTTP handler per domain
├── migrations/                  Skema SQL (dijalankan otomatis & idempotent)
├── scripts/                     Seed data (satuan kerja, RBAC, akun awal)
├── docs/
│   ├── openapi.yaml              Spesifikasi OpenAPI 3.0 lengkap
│   └── swagger-ui/               Swagger UI (di-vendor lokal) — buka /docs
├── web/                          Frontend statis (HTML/CSS/JS, tanpa build step)
│   └── assets/vendor/            Leaflet & Chart.js (di-vendor lokal)
├── vendor/                       Dependensi Go (di-vendor agar build 100% offline)
├── Dockerfile                    Multi-stage build (Go 1.24 alpine → alpine runtime)
├── docker-compose.yml            Service `db` (Postgres 16) + `api`
└── .env.example
```

---

## 3. Arsitektur & Keputusan Desain

### 3.1 Skema Database

Skema mengikuti dokumen sumber (20 entitas, 4 domain: **A. Master & Akses**,
**B. Kasus & Kendaraan**, **C. Interaksi WhatsApp AI**, **D. Monitoring &
Audit**). Dua penyesuaian ditambahkan di luar dokumen sumber, keduanya
diberi komentar `-- ekstensi` pada file migrasi terkait:

1. **`satuan_kerja.induk_id` + nilai `'Polsek'` pada `jenis_satker`** —
   dokumen sumber hanya mencantumkan Polda + 7 Polres/Polresta, sedangkan
   kebutuhan multi-tenant yang diminta ada 3 tingkat (Admin Polda / Admin
   Polres / **Admin Polsek**). Kolom `induk_id` (self-referencing FK)
   ditambahkan untuk membentuk hirarki Polda → Polres → Polsek, dipakai oleh
   `internal/repo/tenant.go` (`ResolveSatkerScope`) untuk membatasi data yang
   terlihat sesuai posisi pengguna di hirarki. Seluruh 45 Polsek jajaran
   (dari 7 Polres/Polresta) sudah di-seed di `scripts/01_seed_satker.sql`.
2. **Tabel `wa_api_key`** — dokumen sumber menjelaskan chatbot WA mengakses
   sistem lewat API, tapi tidak merinci mekanisme kredensialnya. Ditambahkan
   tabel kredensial service-to-service (hash SHA-256 dari API key, key asli
   hanya ditampilkan sekali saat dibuat) agar endpoint `/api/v1/wa/*` punya
   otentikasi terpisah dari sesi JWT pengguna dashboard.
3. **`pengguna.password_hash`** — diperlukan untuk login dashboard
   (bcrypt), tidak eksplisit di skema sumber namun jelas dibutuhkan.

### 3.2 RBAC (Role-Based Access Control)

6 peran (`peran`) × 6 modul (`modul_akses`) → matriks `peran_modul_akses`
dengan level `penuh` / `terbatas` / `ditolak`, persis sesuai tabel "Matriks
Hak Akses" pada dokumen. `terbatas` diartikan sebagai **hanya baca** (method
HTTP mengubah data ditolak, GET tetap diizinkan) — lihat
`internal/middleware/rbac.go`.

RBAC (peran → modul) dan **tenant scoping** (satker → data yang terlihat)
adalah dua sumbu independen:

- **RBAC** menentukan modul apa yang boleh diakses (mis. Operator tidak
  boleh membuka modul Kelola Pengguna).
- **Tenant scope** menentukan data satuan kerja mana yang terlihat (mis.
  Admin Polres Karimun hanya melihat data Polres Karimun + Polsek
  jajarannya, tidak melihat data Polresta Barelang).

Pemetaan peran ↔ istilah "Admin Polda/Polres/Polsek" pada diskusi awal:
**Super Admin** = akun di tingkat Polda (melihat semua), **Admin Polres** =
akun di satuan kerja Polres/Polresta (melihat satker sendiri + Polsek
jajaran), **Operator/Penyidik/Viewer** = biasanya di-assign ke akun tingkat
Polsek (melihat satker sendiri saja). Level akses tetap ditentukan matriks
RBAC yang bisa diubah admin lewat menu Pengaturan Sistem.

### 3.3 Keamanan

- **Password**: bcrypt (`golang.org/x/crypto/bcrypt`).
- **Sesi**: JWT HS256 (implementasi sendiri, hanya stdlib `crypto/hmac` +
  `crypto/sha256`, tanpa dependensi pihak ketiga), disimpan sebagai cookie
  **httpOnly + Secure + SameSite=Strict** — tidak bisa dibaca lewat
  JavaScript (mitigasi XSS token theft), sekaligus tetap mendukung header
  `Authorization: Bearer <token>` untuk kebutuhan API/testing lewat Swagger.
- **NIK pihak terkait**: dienkripsi **AES-256-GCM** di level aplikasi
  sebelum disimpan (kolom `BYTEA`), kunci dari `APP_ENCRYPTION_KEY`. Nilai
  NIK **tidak pernah** dikembalikan utuh oleh API — hanya bentuk tersamar
  (`3271********0001`).
- **API key WA-bot**: disimpan sebagai hash SHA-256, bukan plaintext.
- **Audit trail**: seluruh request yang mengubah data (POST/PUT/PATCH/DELETE)
  dari pengguna terautentikasi otomatis tercatat ke `audit_log`
  (`internal/middleware/audit.go`) — siapa, aktivitas apa, dari IP mana,
  perangkat apa, kapan.
- **Rate limiting**: endpoint login dibatasi 10 percobaan / 5 menit per IP
  (in-memory sliding window) untuk mitigasi brute force.
- **Security headers**: `X-Content-Type-Options`, `X-Frame-Options: DENY`,
  `Content-Security-Policy` ketat (hanya `'self'`, kecuali ubin peta
  OpenStreetMap), `Strict-Transport-Security`.
- **Tanpa dependensi CDN saat runtime**: Leaflet, Chart.js, dan Swagger UI
  di-vendor lokal (`web/assets/vendor/`, `docs/swagger-ui/`) — cocok untuk
  jaringan internal yang aksesnya ke internet publik dibatasi.
- **Build 100% offline**: dependensi Go di-*vendor* (`go mod vendor`), image
  Docker dibangun tanpa perlu mengunduh apa pun dari internet saat
  `docker build`.

### 3.4 Pencarian Cepat ("hitungan detik")

Kolom `kendaraan.no_polisi`, `no_rangka_vin`, `no_mesin` diberi indeks B-tree
(`migrations/0005_indexes_triggers_views.sql`) — inilah jalur kueri utama
yang dipanggil endpoint `/api/v1/wa/search`. Pengujian lokal menunjukkan
waktu respons kueri pencarian di kisaran **1–5 ms** pada data uji.

---

## 4. Integrasi Chatbot WhatsApp AI

Sesuai arahan: *"anggota di lapangan aksesnya lewat API yang terkoneksi ke
knowledge base AI, dia nggak nyentuh-nyentuh database — bagian kita cuma buat
web penginputan dan pencarian data lewat dashboard."*

Proyek ini **tidak** mencakup logika AI/OCR/bot WhatsApp itu sendiri — itu
adalah sistem terpisah (mis. dijalankan tim lain / layanan pihak ketiga).
Yang disediakan di sini adalah **API lengkap dengan dokumentasi Swagger**
yang siap dipanggil oleh sistem tersebut:

1. Buka menu **Pengaturan Sistem → API Key WhatsApp AI**, klik "Buat API
   Key", simpan `raw_key` yang ditampilkan (hanya tampil sekali).
2. Sistem chatbot menyertakan header `X-API-Key: <raw_key>` pada setiap
   panggilan ke `/api/v1/wa/*`.
3. Alur umum per pesan masuk dari petugas via WhatsApp:
   - `POST /api/v1/wa/whitelist/check` — cek nomor pengirim terdaftar &
     aktif (sesuai mekanisme whitelist pada paparan).
   - `POST /api/v1/wa/log-interaksi` — catat pesan masuk (teks/foto
     STNK/foto rangka/foto mesin/voice note).
   - `POST /api/v1/wa/hasil-ocr` — (bila ada) catat hasil OCR/speech-to-text/
     visual matching dari layanan AI.
   - `POST /api/v1/wa/search` — **jalur utama pencarian** (bisa juga
     dipakai langsung tanpa 2 langkah di atas — endpoint ini otomatis
     membuat log interaksi & log pencarian sendiri bila `log_id` tidak
     disertakan). Hasil pencarian otomatis di-scope sesuai satuan kerja
     pemilik nomor WA yang mengirim pesan.
   - `POST /api/v1/wa/respons` — catat balasan yang dikirim AI ke petugas.

Dokumentasi interaktif lengkap (semua endpoint, skema request/response,
tombol "Try it out"): **`http://<host>/docs`**. Spesifikasi mentah:
**`http://<host>/openapi.yaml`**.

---

## 5. Fitur Dashboard

- **Login** NRP + kata sandi.
- **Dashboard Monitoring**: kartu statistik (total kasus, belum/sudah
  terungkap, kendaraan ditemukan/belum), **peta sebaran lokasi** kejadian
  (Leaflet + OpenStreetMap — titik merah = belum terungkap, hijau = sudah
  terungkap/status Selesai) dengan filter rentang tanggal, grafik tren 6
  bulan & distribusi jenis perkara.
- **Data & Pencarian Kasus**: cari laporan berdasarkan kata kunci (No. LP /
  alamat TKP), satuan kerja/wilayah, status kasus, dan **rentang tanggal
  kejadian** (sesuai permintaan: *"aku mau lihat curanmor di batam dalam
  seminggu"* / *"dari tanggal ... sampai tanggal ..."*).
- **Input Laporan Baru & Detail Laporan**: form Laporan Polisi lengkap
  dengan tab Kendaraan, Pihak Terkait (NIK terenkripsi), Barang Bukti,
  Dokumentasi (unggah berkas), Riwayat Penyidikan, dan Riwayat Perubahan
  Status (LP → SP2HP → DPO → Selesai, tercatat otomatis via trigger
  database).
- **Kelola Pengguna**: CRUD petugas, reset password, dan pengelolaan
  whitelist nomor WhatsApp.
- **Pengaturan Sistem** *(Super Admin)*: kelola satuan kerja (termasuk
  menambah Polsek baru), edit matriks RBAC langsung dari UI, kelola API key
  integrasi WhatsApp AI.
- **Laporan & Analitik**: tren 12 bulan, distribusi jenis perkara, laporan
  periodik (harian/mingguan/bulanan), dan audit log.

Seluruh menu otomatis disembunyikan di sisi frontend sesuai matriks RBAC
peran pengguna yang sedang login (selain itu backend tetap menegakkan RBAC
secara independen — penyembunyian menu di frontend murni kenyamanan UX,
bukan mekanisme keamanan).

---

## 6. Pengembangan Lokal (tanpa Docker)

```bash
# Jalankan PostgreSQL 16 lokal, buat database "curanmor_ai"
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/curanmor_ai?sslmode=disable"
export JWT_SECRET="ganti-dengan-32-karakter-acak-xxxxxxxxxx"
export APP_ENCRYPTION_KEY="ganti-dengan-32-karakter-acak-xxxxxxxxxx"
export MIGRATIONS_DIR=./migrations
export SEED_DIR=./scripts
export DOCS_DIR=./docs
export WEB_DIR=./web
export UPLOAD_DIR=./uploads
export COOKIE_SECURE=false   # hanya untuk lokal tanpa HTTPS

go run ./cmd/api
```

Buka `http://localhost:8080`.

---

## 7. Checklist Sebelum Produksi

- [ ] Ganti password akun `00000000` & hapus/ganti nomor whitelist dummy.
- [ ] Set `JWT_SECRET`, `APP_ENCRYPTION_KEY`, `POSTGRES_PASSWORD` ke nilai
      acak unik (jangan pernah pakai nilai contoh).
- [ ] Pasang reverse proxy dengan TLS (HTTPS) di depan container `api`,
      pastikan `COOKIE_SECURE=true`.
- [ ] Batasi `CORS_ALLOWED_ORIGINS` ke domain dashboard resmi (bukan `*`)
      bila frontend disajikan dari domain berbeda.
- [ ] Pertimbangkan menempatkan berkas unggahan (`/app/uploads`) di object
      storage terkelola (mis. MinIO/S3-compatible) untuk deployment
      multi-instance / berskala besar — implementasi saat ini memakai disk
      lokal via Docker volume, cukup untuk single-instance.
- [ ] Backup rutin volume `curanmor_pgdata` (`pg_dump` terjadwal).
- [ ] Review & sesuaikan matriks RBAC di menu Pengaturan Sistem sesuai
      struktur organisasi aktual.
