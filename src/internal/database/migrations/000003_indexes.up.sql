-- Create performance indexes

-- User indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Session indexes
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- API token indexes
CREATE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens(user_id);

-- Organization indexes
CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);

-- Gist indexes
CREATE INDEX IF NOT EXISTS idx_gists_user_id ON gists(user_id);
CREATE INDEX IF NOT EXISTS idx_gists_organization_id ON gists(organization_id);
CREATE INDEX IF NOT EXISTS idx_gists_visibility ON gists(visibility);
CREATE INDEX IF NOT EXISTS idx_gists_created_at ON gists(created_at);
CREATE INDEX IF NOT EXISTS idx_gists_deleted_at ON gists(deleted_at);

-- Gist file indexes
CREATE INDEX IF NOT EXISTS idx_gist_files_gist_id ON gist_files(gist_id);

-- Gist star indexes
CREATE INDEX IF NOT EXISTS idx_gist_stars_gist_id ON gist_stars(gist_id);
CREATE INDEX IF NOT EXISTS idx_gist_stars_user_id ON gist_stars(user_id);

-- Audit log indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);