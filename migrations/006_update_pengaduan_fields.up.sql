-- 006_update_pengaduan_fields.up.sql
ALTER TABLE pengaduan
ADD COLUMN IF NOT EXISTS judul VARCHAR(255);
ADD COLUMN IF NOT EXISTS isi TEXT;
ADD COLUMN IF NOT EXISTS lokasi VARCHAR(255);
ADD COLUMN IF NOT EXISTS tanggal DATE;
ADD COLUMN IF NOT EXISTS keterangan TEXT;
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE pengaduan
ALTER COLUMN status DROP DEFAULT;
ALTER COLUMN status TYPE VARCHAR(50)
USING CASE
	WHEN status::text = 'responded' THEN 'Ditinjau'
	ELSE 'Baru'
END;

ALTER TABLE pengaduan
ALTER COLUMN status SET DEFAULT 'Baru';

UPDATE pengaduan
SET
	foto_url = COALESCE(foto_url, ''),
	judul = COALESCE(judul, 'Tidak ada judul'),
	isi = COALESCE(isi, ''),
	lokasi = COALESCE(lokasi, ''),
	tanggal = COALESCE(tanggal, CURRENT_DATE),
	keterangan = COALESCE(keterangan, ''),
	status = COALESCE(status, 'Baru')
WHERE foto_url IS NULL OR judul IS NULL OR isi IS NULL OR lokasi IS NULL OR tanggal IS NULL OR keterangan IS NULL OR status IS NULL;

ALTER TABLE pengaduan
ALTER COLUMN foto_url SET DEFAULT '',
ALTER COLUMN foto_url SET NOT NULL,
ALTER COLUMN judul SET NOT NULL,
ALTER COLUMN isi SET NOT NULL,
ALTER COLUMN lokasi SET DEFAULT '',
ALTER COLUMN lokasi SET NOT NULL,
ALTER COLUMN tanggal SET DEFAULT CURRENT_DATE,
ALTER COLUMN tanggal SET NOT NULL,
ALTER COLUMN keterangan SET DEFAULT '',
ALTER COLUMN status SET NOT NULL;

-- Remove legacy pengaduan columns that are no longer used by the API
ALTER TABLE pengaduan
DROP COLUMN IF EXISTS pengaduan,
DROP COLUMN IF EXISTS id_app_user,
DROP COLUMN IF EXISTS kontak;

