# Panduan Pengembangan Backend Desa Puundoho

Dokumen ini berisi informasi tentang teknologi, arsitektur, dan daftar API endpoint untuk backend Desa Puundoho.

## Tech Stack

- **Bahasa**: [Go (Golang)](https://go.dev/)
- **Framework Web**: [Gin](https://gin-gonic.com/) (Ringan, cepat, dan mudah untuk routing)
- **Database**: PostgreSQL
- **Driver & Migrasi**: `github.com/lib/pq` dan `golang-migrate/migrate`. File migrasi di-embed langsung ke dalam binary Go sehingga tidak perlu setup manual di server.
- **Autentikasi**: JWT (JSON Web Tokens) menggunakan `golang-jwt/jwt` dan password hashing dengan `bcrypt`.
- **Environment**: `godotenv` untuk meload `.env` saat di environment lokal.

## Struktur File Utama

```text
backend/
├── main.go             # Entry point aplikasi, setup Gin router, dan middleware CORS
├── database.go         # Koneksi ke DB PostgreSQL dan sistem auto-migration
├── auth.go             # Logic Login, manajemen JWT, RBAC Middleware, dan seeder default user
├── articles.go         # Endpoint handler untuk fitur Berita/Artikel (CRUD ke DB)
├── dusun.go            # Endpoint handler untuk fitur Dusun dan Peta (CRUD ke DB)
├── gallery.go          # Endpoint handler untuk fitur Gallery (CRUD ke DB)
├── idm.go              # Endpoint handler untuk fitur IDM (Fetch endpoint pemerintah)
├── listings.go         # Endpoint handler untuk fitur Fasilitas/Listing dan Peta (CRUD ke DB)
├── penduduk.go         # Endpoint handler untuk fitur Penduduk (CRUD ke DB)
├── sdgs.go             # Endpoint handler untuk fitur SDGs (Fetch endpoint pemerintah)
├── docs/               # Generated Swagger documentation (swagger.json, etc.)
├── migrations/         # Folder berisi file migrasi raw SQL (misal: 001_initial_schema.up.sql)
│   └── 001_initial_schema.up.sql # Skema definisi tabel DB pertama kali
├── go.mod / go.sum     # Dependency management Golang
└── .env                # Variabel lingkungan untuk kredensial DB dan ImageKit (tidak di-commit)
```

## Arsitektur & Keamanan

1. **Role-Based Access Control (RBAC)**: Terdapat sistem role-based. Diimplementasikan via JWT payload dan diverifikasi oleh `RoleMiddleware`. Role saat ini: `admin` dan `bendahara`.
2. **Auto-seeding**: Saat awal aplikasi djalankan, backend akan otomatis membuat user default `admin` dan `bendahara` (jika belum ada).
3. **Image Uploading**: Gambar tidak diproses oleh server Go secara langsung dan tidak disimpan sebagai Base64. Server Go hanya memberikan Endpoint `GET /api/imagekit/auth` yang meng-generate token sementara yang bersifat aman. Front-end menggunakan token ini untuk meng-upload ke ImageKit langsung.

## Fitur yang Sedang Dikembangkan (WIP)
- **Visualisasi Data**: Visualisasi data di halaman overview masih kurang.
- **Stunting**: Endpoint dan tabel migrasi sudah ditambahkan ke skema untuk mendata kasus stunting per dusun.
- **Bansos**: Sistem manajemen status penyaluran Bantuan Sosial untuk warga.
- **Pengajuan PPID**: Sistem persuratan dan administrasi request masyarakat ke desa.
- **Pengaduan**: Laporan sudah disupport dengan field kategori dan *foto lampiran*. 

## API Documentation

Backend ini menggunakan **Swagger (OpenAPI 3.0)** untuk dokumentasi API yang interaktif. Anda dapat melihat daftar lengkap endpoint, parameter, dan mencoba request langsung melalui browser.

- **Swagger UI**: [http://localhost:8081/api/swagger/index.html](http://localhost:8081/api/swagger/index.html) ATAU
- **Swagger UI**: [https://desapuundoho.my.id/api/swagger/index.html](https://desapuundoho.my.id/api/swagger/index.html)
- **Spec File**: `backend/docs/swagger.json`

### Cara Update Dokumentasi
Jika Anda mengubah annotation di kode Go, jalankan perintah berikut untuk memperbarui dokumentasi:
```bash
go run github.com/swaggo/swag/cmd/swag init
```

### Protected Endpoints (Admin/Authorized)
*(Memerlukan JWT "Bearer Token" di Header Request)*
- `POST /api/auth/login` - Login admin (bukan protected, return token)
- `POST /api/articles` - Create Artikel
- `POST /api/galeri` - Upload Foto Galeri Baru
- `POST /api/listings` - Add Listing Peta
- `POST /api/penduduk/datasets` - Buat Dataset Tahun Baru
- `POST /api/penduduk/datasets/:id/bulk` - Import Data Excel (JSON)
- `PATCH /api/penduduk/records/:id` - Update baris data penduduk (Excel-like edit)
- `GET /api/imagekit/auth` - Endpoint untuk mengambil Signature Token ImageKit

## Menjalankan Secara Lokal

1. Setup database PostgreSQL dan pastikan username/password sesuai di file `.env`. Untuk lokal pakai neon aja.
2. Buka terminal di direktori `backend`
3. Download dependency: `go mod download`
4. Jalankan server: `go run .`