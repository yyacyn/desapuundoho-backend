ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS kategori VARCHAR(100) NOT NULL DEFAULT 'Umum';

CREATE TABLE IF NOT EXISTS stunting (
    id SERIAL PRIMARY KEY,
    nik_anak VARCHAR(16),
    nama_anak VARCHAR(255) NOT NULL,
    lokasi_dusun VARCHAR(100) NOT NULL,
    tanggal_lahir DATE,
    tinggi_badan DECIMAL(5,2),
    berat_badan DECIMAL(5,2),
    status VARCHAR(50) NOT NULL DEFAULT 'Normal' CHECK (status IN ('Normal', 'Stunting', 'Gizi Buruk', 'Beresiko')),
    tanggal_pemeriksaan DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bansos (
    id SERIAL PRIMARY KEY,
    nama_program VARCHAR(255) NOT NULL,
    nik_penerima VARCHAR(16),
    nama_penerima VARCHAR(255) NOT NULL,
    lokasi_dusun VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Menunggu' CHECK (status IN ('Menunggu', 'Tersalurkan', 'Ditolak')),
    tanggal_penyaluran DATE,
    keterangan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sdgs_desa (
    id SERIAL PRIMARY KEY,
    tahun INT NOT NULL UNIQUE,
    skor_total DECIMAL(5,2) NOT NULL DEFAULT 0,
    file_laporan_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
