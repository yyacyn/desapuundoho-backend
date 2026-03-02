-- 001_create_articles.up.sql
-- Creates the articles table with indexes and auto-update trigger

CREATE TABLE IF NOT EXISTS articles (
    id          SERIAL          PRIMARY KEY,
    title       VARCHAR(255)    NOT NULL,
    slug        VARCHAR(255)    NOT NULL UNIQUE,
    content     TEXT            NOT NULL,
    excerpt     VARCHAR(500),
    cover_image VARCHAR(512),
    author      VARCHAR(100)    NOT NULL DEFAULT 'Admin',
    status      VARCHAR(20)     NOT NULL DEFAULT 'draft'
                                CHECK (status IN ('draft', 'published')),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_slug   ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);

-- Auto-update updated_at whenever a row is modified
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
