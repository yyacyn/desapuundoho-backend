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
├── listings.go         # Endpoint handler untuk fitur Fasilitas/Listing dan Peta (CRUD ke DB)
├── migrations/         # Folder berisi file migrasi raw SQL (misal: 001_initial_schema.up.sql)
│   └── 001_initial_schema.up.sql # Skema definisi tabel DB pertama kali
├── go.mod / go.sum     # Dependency management Golang
└── .env                # Variabel lingkungan untuk kredensial DB dan ImageKit (tidak di-commit)
```

## Arsitektur & Keamanan

1. **Role-Based Access Control (RBAC)**: Terdapat sistem role-based. Diimplementasikan via JWT payload dan diverifikasi oleh `RoleMiddleware`. Role saat ini: `admin` dan `bendahara`.
2. **Auto-seeding**: Saat awal aplikasi djalankan, backend akan otomatis membuat user default `admin` dan `bendahara` (jika belum ada).
3. **Image Uploading**: Gambar tidak diproses oleh server Go secara langsung dan tidak disimpan sebagai Base64. Server Go hanya memberikan Endpoint `GET /api/imagekit/auth` yang meng-generate token sementara yang bersifat aman. Front-end menggunakan token ini untuk meng-upload ke ImageKit langsung.

## API Documentation (Detailed)

Semua route diawali dengan base URL: `http://localhost:8081`

### 1. GET Articles
Retrieve a list of published articles.
- **URL:** `/api/articles`
- **Method:** `GET`

### 2. GET Galeri
Retrieve a list of gallery items with images.
- **URL:** `/api/galeri`
- **Method:** `GET`

### 3. GET Listings
Retrieve a list of location listings.
- **URL:** `/api/listings`
- **Method:** `GET`

### 4. GET Population Stats
Retrieve population statistics for a given dataset.
- **URL:** `/api/penduduk/datasets/:id/stats`
- **Method:** `GET`

#### Example Response — `200 OK`
```json
{
  "age_range": { "0-5": 70, "6-12": 131, "18-59": 729, ... },
  "dusun": { "Dusun 1": 464, "Dusun 2": 211, ... },
  "education": { "SD Sederajat": 390, "SMP Sederajat": 180, ... },
  "gender": { "Laki-laki": 571, "Perempuan": 568 },
  "job": { "Petani": 211, "IRT": 251, "Wiraswasta": 55, ... },
  "marriage": { "Belum Kawin": 569, "Kawin": 509, ... },
  "religion": { "Islam": 1123, "Kristen": 15 ... }
}
```

> Untuk dokumentasi lengkap beserta contoh JSON tiap endpoint, silakan rujuk ke file `coll_doc.md`.

### Protected Endpoints (Admin/Authorized)
- `POST /api/auth/login` - Login admin
- `POST /api/articles` - Create Artikel
- `POST /api/galeri` - Upload Foto Galeri Baru
- `POST /api/listings` - Add Listing Peta
- `POST /api/penduduk/datasets` - Buat Dataset Tahun Baru
- `POST /api/penduduk/datasets/:id/bulk` - Import Data Excel (JSON)
- `PATCH /api/penduduk/records/:id` - Update baris data penduduk (Excel-like edit)

## Menjalankan Secara Lokal

1. Setup database PostgreSQL dan pastikan username/password sesuai di file `.env`. Untuk lokal pakai neon aja.
2. Buka terminal di direktori `backend`
3. Download dependency: `go mod download`
4. Jalankan server: `go run .`

## Proses Deploy (ini gausah dipikirin biar aku aja yang deploy)

Untuk deploy, build service menjadi sebuah file biner / *executable* tunggal.
Untuk server hosting berbasis Linux (seperti VPS Ubuntu / cPanel Terminal):

```bash
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server
```

Lalu angkat file bernama `server` ke direktori root hosting dan jalankan service-nya.
