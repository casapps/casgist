-- Rollback initial schema

-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_user_id;

DROP INDEX IF EXISTS idx_gist_stars_user_id;
DROP INDEX IF EXISTS idx_gist_stars_gist_id;

DROP INDEX IF EXISTS idx_gist_files_gist_id;

DROP INDEX IF EXISTS idx_gists_deleted_at;
DROP INDEX IF EXISTS idx_gists_created_at;
DROP INDEX IF EXISTS idx_gists_visibility;
DROP INDEX IF EXISTS idx_gists_organization_id;
DROP INDEX IF EXISTS idx_gists_user_id;

DROP INDEX IF EXISTS idx_organizations_deleted_at;
DROP INDEX IF EXISTS idx_organizations_name;

DROP INDEX IF EXISTS idx_api_tokens_user_id;
DROP INDEX IF EXISTS idx_api_tokens_token_hash;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;

DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

-- Drop tables in reverse order
DROP TABLE IF EXISTS passkeys;
DROP TABLE IF EXISTS import_jobs;
DROP TABLE IF EXISTS backup_jobs;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS system_configs;
DROP TABLE IF EXISTS user_follows;
DROP TABLE IF EXISTS gist_comments;
DROP TABLE IF EXISTS gist_stars;
DROP TABLE IF EXISTS gist_files;
DROP TABLE IF EXISTS gists;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS users;