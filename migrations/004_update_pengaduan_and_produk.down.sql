-- 004_update_pengaduan_and_produk.down.sql

-- 1. Restore pengaduan table attributes
ALTER TABLE pengaduan DROP COLUMN IF EXISTS nama;
ALTER TABLE pengaduan DROP COLUMN IF EXISTS kontak;
ALTER TABLE pengaduan DROP COLUMN IF EXISTS kategori;
ALTER TABLE pengaduan ADD COLUMN id_app_user INT REFERENCES app_users(id) ON DELETE CASCADE;

-- 2. Restore rating to produk_desa
ALTER TABLE produk_desa ADD COLUMN rating DECIMAL(2,1) NOT NULL DEFAULT 0.0;
