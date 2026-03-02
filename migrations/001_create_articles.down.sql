-- 001_create_articles.down.sql
-- Rolls back: drops trigger, function, indexes, and table

DROP TRIGGER  IF EXISTS set_articles_updated_at ON articles;
DROP FUNCTION IF EXISTS update_updated_at();
DROP INDEX    IF EXISTS idx_articles_status;
DROP INDEX    IF EXISTS idx_articles_slug;
DROP TABLE    IF EXISTS articles;
