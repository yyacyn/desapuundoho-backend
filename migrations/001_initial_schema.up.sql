-- 001_create_initial_schema.up.sql
-- Merges articles, admin_users, and the rest of the village schema into a single initial migration

-- 1. Articles
CREATE TABLE IF NOT EXISTS articles (
    id          SERIAL          PRIMARY KEY,
    title       VARCHAR(255)    NOT NULL,
    slug        VARCHAR(255)    NOT NULL UNIQUE,
    content     TEXT            NOT NULL,
    excerpt     VARCHAR(500),
    cover_image VARCHAR(512),
    author      VARCHAR(100)    NOT NULL DEFAULT 'Admin',
    category    VARCHAR(100)    NOT NULL DEFAULT 'Umum',
    status      VARCHAR(20)     NOT NULL DEFAULT 'draft'
                                CHECK (status IN ('draft', 'published')),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_slug   ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

-- 2. Admin Users
CREATE TABLE IF NOT EXISTS admin_users (
    id            SERIAL        PRIMARY KEY,
    username      VARCHAR(100)  NOT NULL UNIQUE,
    password_hash TEXT          NOT NULL,
    nama          VARCHAR(255),
    role          VARCHAR(50)   NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username);

-- 3. Web App Users (For Residents)
CREATE TABLE IF NOT EXISTS app_users (
    id SERIAL PRIMARY KEY,
    nik VARCHAR(16) UNIQUE, 
    nama VARCHAR(255) NOT NULL,
    alamat TEXT NOT NULL,
    telp VARCHAR(20) NOT NULL,
    jenis_kelamin VARCHAR(1) NOT NULL CHECK (jenis_kelamin IN ('L', 'P')),
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Official Resident Data (Penduduk & Datasets)
CREATE TABLE IF NOT EXISTS penduduk_datasets (
    id SERIAL PRIMARY KEY,
    tahun INT NOT NULL UNIQUE,
    nama_file VARCHAR(255),
    total_records INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS penduduk (
    id SERIAL PRIMARY KEY,
    dataset_id INT REFERENCES penduduk_datasets(id) ON DELETE CASCADE,
    nik VARCHAR(16) NOT NULL,
    no_kk VARCHAR(16) NOT NULL,
    nama VARCHAR(255) NOT NULL,
    jenis_kelamin VARCHAR(1) NOT NULL CHECK (jenis_kelamin IN ('L', 'P')),
    status_kawin VARCHAR(50) NOT NULL,
    tempat_lahir VARCHAR(100) NOT NULL,
    tanggal_lahir DATE NOT NULL,
    agama VARCHAR(50) NOT NULL,
    pend_terakhir VARCHAR(100),
    pekerjaan VARCHAR(100),
    bisa_baca BOOLEAN NOT NULL DEFAULT FALSE,
    kewarganegaraan VARCHAR(50) NOT NULL DEFAULT 'WNI',
    alamat TEXT NOT NULL,
    kedudukan_keluarga VARCHAR(100),
    UNIQUE (nik, dataset_id)
);

CREATE INDEX IF NOT EXISTS idx_penduduk_nik ON penduduk(nik);
CREATE INDEX IF NOT EXISTS idx_penduduk_dataset_id ON penduduk(dataset_id);

-- 5. Pengaduan (Complaints)
CREATE TABLE IF NOT EXISTS pengaduan (
    id SERIAL PRIMARY KEY,
    id_app_user INT NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    pengaduan TEXT NOT NULL,
    foto_url VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'menunggu',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pengaduan_status ON pengaduan(status);

-- 6. Pengajuan (Requests for Official Letters)
CREATE TABLE IF NOT EXISTS pengajuan (
    id BIGSERIAL PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    jenis_pengajuan VARCHAR(100) NOT NULL,
    biaya BIGINT NOT NULL DEFAULT 0,
    kontak VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pengajuan_status ON pengajuan(status);

-- 7. Galeri & Listings
CREATE TABLE IF NOT EXISTS galeri (
    id SERIAL PRIMARY KEY,
    images JSONB NOT NULL DEFAULT '[]'::jsonb,
    caption TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS listings (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    koordinat VARCHAR(255) NOT NULL,
    image_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. PPID & Bagan Organisasi
CREATE TABLE IF NOT EXISTS ppid (
    id SERIAL PRIMARY KEY,
    id_parent INT REFERENCES ppid(id) ON DELETE RESTRICT,
    judul VARCHAR(255) NOT NULL,
    deskripsi TEXT,
    file_dokumen_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bagan_organisasi (
    id SERIAL PRIMARY KEY,
    image_url JSONB NOT NULL DEFAULT '[]'::jsonb,
    caption TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. IDM (Indeks Desa Membangun)
CREATE TABLE IF NOT EXISTS idm (
    id SERIAL PRIMARY KEY,
    tahun INT NOT NULL UNIQUE,
    skor_idm DECIMAL(10,4) NOT NULL DEFAULT 0,
    status VARCHAR(100) NOT NULL,
    target VARCHAR(100) NOT NULL,
    skor_minimal DECIMAL(10,4) NOT NULL DEFAULT 0,
    penambahan DECIMAL(10,4) NOT NULL DEFAULT 0,
    skor_iks DECIMAL(10,4) NOT NULL DEFAULT 0,
    skor_ike DECIMAL(10,4) NOT NULL DEFAULT 0,
    skor_ikl DECIMAL(10,4) NOT NULL DEFAULT 0,
    file_dokumen_url VARCHAR(500)
);

-- 10. BUMDes / Produk Desa
CREATE TABLE IF NOT EXISTS produk_desa (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    deskripsi TEXT,
    harga BIGINT NOT NULL DEFAULT 0,
    rating DECIMAL(2,1) NOT NULL DEFAULT 0.0,
    kontak VARCHAR(100) NOT NULL,
    image_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. APBD Desa (Village Budget)
CREATE TABLE IF NOT EXISTS apbd_desa (
    id SERIAL PRIMARY KEY,
    tahun INT NOT NULL UNIQUE,
    total_pendapatan BIGINT NOT NULL DEFAULT 0,
    total_pengeluaran BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS apbd_pendapatan (
    id SERIAL PRIMARY KEY,
    id_apbd INT NOT NULL REFERENCES apbd_desa(id) ON DELETE CASCADE,
    kategori VARCHAR(255) NOT NULL,
    jumlah BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS apbd_pengeluaran (
    id SERIAL PRIMARY KEY,
    id_apbd INT NOT NULL REFERENCES apbd_desa(id) ON DELETE CASCADE,
    bidang VARCHAR(255) NOT NULL,
    jumlah BIGINT NOT NULL DEFAULT 0
);


