-- 005_drop_app_users_and_update_pengaduan.down.sql

-- 1. Restore app_users table
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

DO $$
DECLARE
    current_udt text;
BEGIN
    SELECT udt_name
    INTO current_udt
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'pengaduan'
      AND column_name = 'status';

    IF current_udt = 'pengaduan_status_enum' THEN
        EXECUTE $sql$
            ALTER TABLE pengaduan
            ALTER COLUMN status DROP DEFAULT,
            ALTER COLUMN status TYPE VARCHAR(50)
            USING CASE
                WHEN status::text = 'responded' THEN 'responded'
                ELSE 'menunggu'
            END,
            ALTER COLUMN status SET DEFAULT 'menunggu'
        $sql$;
    ELSE
        EXECUTE 'ALTER TABLE pengaduan ALTER COLUMN status SET DEFAULT ''menunggu''';
    END IF;
END $$;

-- 2. Restore kontak and move values back from nomor_telp when available
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS kontak VARCHAR(100) NOT NULL DEFAULT '';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'pengaduan' AND column_name = 'nomor_telp'
    ) THEN
        EXECUTE 'UPDATE pengaduan SET kontak = COALESCE(NULLIF(nomor_telp, ''''''), kontak)';
    END IF;
END $$;

-- 3. Remove columns added in this migration
ALTER TABLE pengaduan DROP COLUMN IF EXISTS email;
ALTER TABLE pengaduan DROP COLUMN IF EXISTS nomor_telp;
DROP TYPE IF EXISTS pengaduan_status_enum;