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

Semua route diawali dengan base URL: `http://localhost:8081`

### 1. GET Articles
Retrieve a list of published articles.
- **Method:** `GET`
- **URL:** `/api/articles`

### 2. GET Galeri
Retrieve a list of gallery items with images.
- **Method:** `GET`
- **URL:** `/api/galeri`

### 3. GET Listings
Retrieve a list of location listings.
- **Method:** `GET`
- **URL:** `/api/listings`

### 4. GET Population Stats
Retrieve population statistics for a given dataset, mapping them to regions and demography.
- **Method:** `GET`
- **URL:** `/api/penduduk/datasets/:id/stats`
*(Ganti `:id` dengan referensial ID tahun dataset).*

#### Example Response — `200 OK`
```json
{
  "age_range": { "18-59": 729, "60+": 100, "...": "..." },
  "age_by_dusun": { "18-59": { "Dusun 1": 304, "Dusun 2": 129, "...": "..." } },
  "dusun": { "Dusun 1": 464, "Dusun 2": 211, "...": "..." },
  "education": { "SD Sederajat": 390, "SMP Sederajat": 180, "...": "..." },
  "gender": { "Laki-laki": 571, "Perempuan": 568 },
  "gender_by_dusun": { "Laki-laki": { "Dusun 1": 244, "Dusun 2": 95, "...": "..." } },
  "job": { "Petani": 211, "IRT": 251, "Wiraswasta": 55, "...": "..." },
  "marriage": { "Belum Kawin": 569, "Kawin": 509, "...": "..." },
  "religion": { "Islam": 1123, "Kristen": 15, "...": "..." },
  "religion_by_dusun": { "Islam": { "Dusun 1": 462, "Dusun 2": 211, "...": "..." } }
}
```

### 5. GET Dusun Boundaries
Retrieve mapping data, boundary colors, and GeoJSON shapes for each Dusun.
- **Method:** `GET`
- **URL:** `/api/dusun`

#### Example Response — `200 OK`
```json
{
    "dusun": [
        {
            "id": 1,
            "nama_dusun": "Dusun 1",
            "warna": "#298064",
            "geojson_data": "{\"type\": \"Polygon\", \"coordinates\": [[[121.09, -3.10]...]]}"
        }
    ]
}
```

### 6. GET SDGs Data
Retrieve live SDGs (Sustainable Development Goals) score data from Kemendesa API.
- **Method:** `GET`
- **URL:** `/api/sdgs`

#### Example Response — `200 OK`
```json
{
    "average": "46.58",
    "data": [
        {
            "goals": 1,
            "title": "Desa Tanpa Kemiskinan",
            "score": 51.63
        },
        {
            "goals": 2,
            "title": "Desa Tanpa Kelaparan",
            "score": 62.5
        }
    ],
    "total_desa": 1
}
```

### 7. GET IDM Data
Retrieve live IDM (Indeks Desa Membangun) data from Kemendesa API. 
**Note:** You can include an optional `tahun` query parameter to filter by year. If the year returned no data, the API will output unavailable data gracefully.
- **Method:** `GET`
- **URL:** `/api/idm` or `/api/idm?tahun=2024`

#### Example Response — `200 OK`
```json
{
    "status": 200,
    "error": false,
    "mapData": {
        "SUMMARIES": {
            "SKOR_SAAT_INI": 0.6998,
            "STATUS": "BERKEMBANG",
            "TARGET_STATUS": "MAJU",
            "TAHUN": 2024
        },
        "ROW": [
            {
                "NO": 1,
                "INDIKATOR": "Skor Akses Sarkes",
                "SKOR": 5,
                "KETERANGAN": "Waktu tempuh dari ≤ 30  Menit"
            }
        ],
        "IDENTITAS": [
            {
                "nama_desa": "PUUNDOHO",
                "nama_kecamatan": "PAKUE UTARA"
            }
        ]
    }
}
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
