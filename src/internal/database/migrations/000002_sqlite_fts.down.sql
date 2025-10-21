-- Rollback SQLite FTS5 virtual table

-- Drop triggers
DROP TRIGGER IF EXISTS gists_fts_soft_delete;
DROP TRIGGER IF EXISTS gists_fts_delete;
DROP TRIGGER IF EXISTS gists_fts_update;
DROP TRIGGER IF EXISTS gists_fts_insert;

-- Drop FTS table
DROP TABLE IF EXISTS gists_fts;