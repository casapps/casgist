-- Drop performance indexes

-- Audit log indexes
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_user_id;

-- Gist star indexes
DROP INDEX IF EXISTS idx_gist_stars_user_id;
DROP INDEX IF EXISTS idx_gist_stars_gist_id;

-- Gist file indexes
DROP INDEX IF EXISTS idx_gist_files_gist_id;

-- Gist indexes
DROP INDEX IF EXISTS idx_gists_deleted_at;
DROP INDEX IF EXISTS idx_gists_created_at;
DROP INDEX IF EXISTS idx_gists_visibility;
DROP INDEX IF EXISTS idx_gists_organization_id;
DROP INDEX IF EXISTS idx_gists_user_id;

-- Organization indexes
DROP INDEX IF EXISTS idx_organizations_deleted_at;
DROP INDEX IF EXISTS idx_organizations_name;

-- API token indexes
DROP INDEX IF EXISTS idx_api_tokens_user_id;
DROP INDEX IF EXISTS idx_api_tokens_token_hash;

-- Session indexes
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;

-- User indexes
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;