-- 006_update_pengaduan_fields.down.sql

-- Remove new columns from pengaduan table
ALTER TABLE pengaduan 
DROP COLUMN IF EXISTS judul,
DROP COLUMN IF EXISTS isi,
DROP COLUMN IF EXISTS lokasi,
DROP COLUMN IF EXISTS tanggal,
DROP COLUMN IF EXISTS keterangan,
DROP COLUMN IF EXISTS status;
