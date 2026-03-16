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

## API Routes Documentation

Semua route diawali dengan base URL: `http://localhost:8081` (atau port yang disetting).

### **Public Endpoints** (Tanpa Autentikasi)

- `POST /api/auth/login`
  - Body: `{ "username": "...", "password": "..." }`
  - Mereturn JWT token dan info user/role jika berhasil.
- `GET  /api/articles`
  - Mengambil daftar semua artikel. Mendukung query param `?status=published`.
- `GET  /api/articles/:id`
  - Mengambil detail 1 artikel berdasarkan ID.
- `GET  /api/listings`
  - Mengambil daftar letak/fasilitas desa.
- `GET  /api/listings/:id`
  - Mengambil detail info 1 fasilitas spesifik.

### **Protected Endpoints** (Wajib Kirim Header `Authorization: Bearer <token>`)

- `GET  /api/auth/me`
  - Mengembalikan informasi identitas user & role yang login saat ini berdasarkan verifikasi Token.
- `GET  /api/imagekit/auth`
  - Endpoint utilitas yang mereturn token JWT rahasia sekali pakai milik ImageKit (signature) untuk memberikan akses upload sementara.

### **Protected Endpoints + Admin Role** (Wajib Login & Wajib Berperan Sebagai 'Admin')

- `POST   /api/articles` (Create Artikel Baru)
- `PUT    /api/articles/:id` (Update Edit Artikel)
- `DELETE /api/articles/:id` (Hapus Artikel)
- `POST   /api/listings` (Add Listing Peta Baru)
- `PUT    /api/listings/:id` (Update Listing)
- `DELETE /api/listings/:id` (Hapus Listing)

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
