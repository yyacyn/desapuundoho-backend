-- 004_update_pengaduan_and_produk.up.sql

-- 1. Modify pengaduan table (Remove id_app_user, add nama, kontak, kategori)
ALTER TABLE pengaduan DROP COLUMN IF EXISTS id_app_user;
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS nama VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS kontak VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS kategori VARCHAR(100) NOT NULL DEFAULT 'Umum';

-- 2. Modify produk_desa table (Remove rating)
ALTER TABLE produk_desa DROP COLUMN IF EXISTS rating;
