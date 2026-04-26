-- 005_drop_app_users_and_update_pengaduan.up.sql

-- 1. Ensure legacy relation column is removed from pengaduan if it still exists
ALTER TABLE pengaduan DROP COLUMN IF EXISTS id_app_user;

DO $$
DECLARE
    current_udt text;
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type
        WHERE typname = 'pengaduan_status_enum'
    ) THEN
        CREATE TYPE pengaduan_status_enum AS ENUM ('submitted', 'responded');
    END IF;

    SELECT udt_name
    INTO current_udt
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'pengaduan'
      AND column_name = 'status';

    IF current_udt IS NOT NULL AND current_udt <> 'pengaduan_status_enum' THEN
        EXECUTE $sql$
            ALTER TABLE pengaduan
            ALTER COLUMN status DROP DEFAULT,
            ALTER COLUMN status TYPE pengaduan_status_enum
            USING CASE
                WHEN status = 'responded' THEN 'responded'::pengaduan_status_enum
                ELSE 'submitted'::pengaduan_status_enum
            END,
            ALTER COLUMN status SET DEFAULT 'submitted'
        $sql$;
    ELSE
        EXECUTE 'ALTER TABLE pengaduan ALTER COLUMN status SET DEFAULT ''submitted''';
    END IF;
END $$;

-- 2. Keep nama on pengaduan (for compatibility if this migration runs independently)
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS nama VARCHAR(255) NOT NULL DEFAULT '';

-- 3. Replace kontak with nomor_telp while preserving data
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS nomor_telp VARCHAR(100) NOT NULL DEFAULT '';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'pengaduan' AND column_name = 'kontak'
    ) THEN
        EXECUTE 'UPDATE pengaduan SET nomor_telp = COALESCE(NULLIF(kontak, ''''''), nomor_telp)';
    END IF;
END $$;

ALTER TABLE pengaduan DROP COLUMN IF EXISTS kontak;

-- 4. Add email to pengaduan
ALTER TABLE pengaduan ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT '';

-- 5. Drop resident app users table
DROP TABLE IF EXISTS app_users CASCADE;