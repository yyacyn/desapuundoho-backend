-- 001_create_initial_schema.down.sql
-- Rolls back all tables created in the initial schema Migration

DROP INDEX IF EXISTS idx_pengaduan_status;
DROP INDEX IF EXISTS idx_pengajuan_status;
DROP INDEX IF EXISTS idx_penduduk_nik;

DROP TABLE IF EXISTS apbd_pengeluaran;
DROP TABLE IF EXISTS apbd_pendapatan;
DROP TABLE IF EXISTS apbd_desa CASCADE;
DROP TABLE IF EXISTS produk_desa;
DROP TABLE IF EXISTS bagan_organisasi;
DROP TABLE IF EXISTS ppid CASCADE;
DROP TABLE IF EXISTS pengajuan;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS galeri;
DROP TABLE IF EXISTS pengaduan CASCADE;
DROP TABLE IF EXISTS penduduk;
DROP TABLE IF EXISTS app_users CASCADE;
DROP TABLE IF EXISTS admin_users CASCADE;

DROP TRIGGER IF EXISTS set_articles_updated_at ON articles;
DROP FUNCTION IF EXISTS update_updated_at();
DROP INDEX IF EXISTS idx_articles_status;
DROP INDEX IF EXISTS idx_articles_slug;
DROP TABLE IF EXISTS articles CASCADE;
