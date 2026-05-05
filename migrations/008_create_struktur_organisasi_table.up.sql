-- 008_create_struktur_organisasi_table.up.sql
CREATE TABLE IF NOT EXISTS struktur_organisasi (
    id SERIAL PRIMARY KEY,
    image_url VARCHAR(500) NOT NULL,
    caption TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_struktur_organisasi_created_at ON struktur_organisasi(created_at DESC);
