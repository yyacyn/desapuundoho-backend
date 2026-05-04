-- 007_create_pengajuan_table.down.sql

UPDATE pengajuan
SET status = 'pending'
WHERE status IS NOT NULL;

ALTER TABLE pengajuan
ALTER COLUMN status SET DEFAULT 'pending';

ALTER TABLE pengajuan
DROP COLUMN IF EXISTS judul,
DROP COLUMN IF EXISTS isi,
DROP COLUMN IF EXISTS dokumen_url,
DROP COLUMN IF EXISTS kategori,
DROP COLUMN IF EXISTS nomor_telp,
DROP COLUMN IF EXISTS email,
DROP COLUMN IF EXISTS lokasi,
DROP COLUMN IF EXISTS tanggal,
DROP COLUMN IF EXISTS keterangan,
DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_pengajuan_status;
DROP INDEX IF EXISTS idx_pengajuan_created_at;

