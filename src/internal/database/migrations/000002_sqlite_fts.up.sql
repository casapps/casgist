-- Create SQLite FTS5 virtual table for search
-- This migration only runs for SQLite databases

-- Create FTS5 virtual table for gists search
CREATE VIRTUAL TABLE IF NOT EXISTS gists_fts USING fts5(
    gist_id UNINDEXED,
    title,
    description,
    content,
    username,
    filename,
    language,
    tags,
    tokenize='porter unicode61'
);

-- Create triggers to keep FTS table in sync
-- Insert trigger
CREATE TRIGGER IF NOT EXISTS gists_fts_insert 
AFTER INSERT ON gists
BEGIN
    INSERT INTO gists_fts (gist_id, title, description, username)
    SELECT 
        NEW.id,
        NEW.title,
        NEW.description,
        COALESCE((SELECT username FROM users WHERE id = NEW.user_id), '')
    WHERE NEW.deleted_at IS NULL;
END;

-- Update trigger
CREATE TRIGGER IF NOT EXISTS gists_fts_update
AFTER UPDATE ON gists
BEGIN
    DELETE FROM gists_fts WHERE gist_id = NEW.id;
    INSERT INTO gists_fts (gist_id, title, description, username)
    SELECT 
        NEW.id,
        NEW.title,
        NEW.description,
        COALESCE((SELECT username FROM users WHERE id = NEW.user_id), '')
    WHERE NEW.deleted_at IS NULL;
END;

-- Delete trigger
CREATE TRIGGER IF NOT EXISTS gists_fts_delete
AFTER DELETE ON gists
BEGIN
    DELETE FROM gists_fts WHERE gist_id = OLD.id;
END;

-- Soft delete trigger
CREATE TRIGGER IF NOT EXISTS gists_fts_soft_delete
AFTER UPDATE OF deleted_at ON gists
WHEN NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL
BEGIN
    DELETE FROM gists_fts WHERE gist_id = NEW.id;
END;