-- 007_create_pengajuan_table.up.sql
ALTER TABLE pengajuan
ADD COLUMN IF NOT EXISTS judul VARCHAR(255),
ADD COLUMN IF NOT EXISTS isi TEXT,
ADD COLUMN IF NOT EXISTS dokumen_url VARCHAR(500),
ADD COLUMN IF NOT EXISTS kategori VARCHAR(100),
ADD COLUMN IF NOT EXISTS nomor_telp VARCHAR(20),
ADD COLUMN IF NOT EXISTS email VARCHAR(100),
ADD COLUMN IF NOT EXISTS lokasi VARCHAR(255),
ADD COLUMN IF NOT EXISTS tanggal DATE,
ADD COLUMN IF NOT EXISTS keterangan TEXT,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

UPDATE pengajuan
SET
    judul = COALESCE(NULLIF(judul, ''), NULLIF(jenis_pengajuan, ''), NULLIF(nama, ''), 'Pengajuan'),
    isi = COALESCE(NULLIF(isi, ''), ''),
    dokumen_url = COALESCE(dokumen_url, ''),
    status = CASE
        WHEN status IS NULL OR status = 'pending' THEN 'Baru'
        ELSE status
    END,
    kategori = COALESCE(NULLIF(kategori, ''), NULLIF(jenis_pengajuan, ''), 'Umum'),
    nomor_telp = COALESCE(NULLIF(nomor_telp, ''), NULLIF(kontak, ''), ''),
    email = COALESCE(NULLIF(email, ''), ''),
    lokasi = COALESCE(NULLIF(lokasi, ''), ''),
    tanggal = COALESCE(tanggal, created_at::date, CURRENT_DATE),
    keterangan = COALESCE(keterangan, '')
WHERE judul IS NULL
   OR isi IS NULL
   OR dokumen_url IS NULL
   OR status IS NULL
   OR kategori IS NULL
   OR nomor_telp IS NULL
   OR email IS NULL
   OR lokasi IS NULL
   OR tanggal IS NULL
   OR keterangan IS NULL;

ALTER TABLE pengajuan
ALTER COLUMN judul SET DEFAULT 'Pengajuan',
ALTER COLUMN judul SET NOT NULL,
ALTER COLUMN isi SET DEFAULT '',
ALTER COLUMN isi SET NOT NULL,
ALTER COLUMN dokumen_url SET DEFAULT '',
ALTER COLUMN dokumen_url SET NOT NULL,
ALTER COLUMN status SET DEFAULT 'Baru',
ALTER COLUMN status SET NOT NULL,
ALTER COLUMN kategori SET DEFAULT 'Umum',
ALTER COLUMN kategori SET NOT NULL,
ALTER COLUMN nama SET DEFAULT '',
ALTER COLUMN nama SET NOT NULL,
ALTER COLUMN nomor_telp SET DEFAULT '',
ALTER COLUMN nomor_telp SET NOT NULL,
ALTER COLUMN email SET DEFAULT '',
ALTER COLUMN email SET NOT NULL,
ALTER COLUMN lokasi SET DEFAULT '',
ALTER COLUMN lokasi SET NOT NULL,
ALTER COLUMN tanggal SET DEFAULT CURRENT_DATE,
ALTER COLUMN tanggal SET NOT NULL,
ALTER COLUMN keterangan SET DEFAULT '',
ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_pengajuan_status ON pengajuan(status);
CREATE INDEX IF NOT EXISTS idx_pengajuan_created_at ON pengajuan(created_at DESC);


