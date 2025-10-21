# CasGists Migration Guide

## Table of Contents
- [Migrating from OpenGist](#migrating-from-opengist)
- [Importing from GitHub](#importing-from-github)
- [Database Migrations](#database-migrations)
- [Version Upgrades](#version-upgrades)
- [Data Export/Import](#data-exportimport)

## Migrating from OpenGist

### Prerequisites
- Access to OpenGist database
- CasGists instance running
- Sufficient storage for migration

### Migration Process

#### 1. Prepare Migration
```bash
# Test connection to OpenGist database
./casgists migrate opengist test \
  --source-db "postgres://user:pass@opengist-host/opengist"

# Analyze migration scope
./casgists migrate opengist analyze \
  --source-db "postgres://user:pass@opengist-host/opengist" \
  --output migration-report.json
```

#### 2. Execute Migration
```bash
# Full migration with all options
./casgists migrate opengist \
  --source-db "postgres://user:pass@opengist-host/opengist" \
  --source-files /path/to/opengist/files \
  --map-users \
  --reset-passwords \
  --preserve-dates \
  --batch-size 100
```

#### 3. Migration Options
- `--map-users`: Create user mapping file
- `--reset-passwords`: Require password reset
- `--preserve-dates`: Keep original timestamps
- `--skip-ssh-keys`: Don't migrate SSH keys
- `--dry-run`: Preview without changes

#### 4. Post-Migration
```bash
# Verify migration
./casgists migrate verify \
  --report migration-report.json

# Send password reset emails
./casgists migrate notify-users \
  --template password-reset
```

### User Mapping

Create a user mapping file for custom username changes:

```json
{
  "user_mappings": {
    "old_username1": "new_username1",
    "old_username2": "new_username2"
  },
  "email_mappings": {
    "old@email.com": "new@email.com"
  }
}
```

Use with: `--user-mapping mapping.json`

### Troubleshooting OpenGist Migration

#### Connection Issues
```bash
# Test with psql
psql -h opengist-host -U user -d opengist -c "SELECT version();"

# Check network connectivity
nc -zv opengist-host 5432
```

#### Permission Issues
```sql
-- Grant read permissions (run on OpenGist)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO migration_user;
```

#### Large Migrations
```bash
# Split into batches
./casgists migrate opengist \
  --source-db "..." \
  --batch-size 50 \
  --offset 0 \
  --limit 1000

# Resume from checkpoint
./casgists migrate opengist \
  --resume-from checkpoint.json
```

## Importing from GitHub

### Prerequisites
- GitHub Personal Access Token
- Sufficient API rate limit
- Storage for gist content

### Generate GitHub Token

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Click "Generate new token (classic)"
3. Select scopes:
   - `gist` (required)
   - `read:user` (for user info)
4. Save token securely

### Import Process

#### 1. Basic Import
```bash
# Import all your gists
./casgists import github \
  --token YOUR_GITHUB_TOKEN \
  --username YOUR_USERNAME
```

#### 2. Advanced Import
```bash
# Import with all options
./casgists import github \
  --token YOUR_GITHUB_TOKEN \
  --username YOUR_USERNAME \
  --include-private \
  --include-starred \
  --include-forks \
  --preserve-dates \
  --download-comments \
  --auto-tags
```

#### 3. Organization Import
```bash
# Import organization gists
./casgists import github-org \
  --token YOUR_GITHUB_TOKEN \
  --org ORGANIZATION_NAME \
  --create-org \
  --map-members
```

#### 4. Selective Import
```bash
# Import specific gists
./casgists import github \
  --token YOUR_GITHUB_TOKEN \
  --gist-ids gist1,gist2,gist3

# Import by criteria
./casgists import github \
  --token YOUR_GITHUB_TOKEN \
  --username YOUR_USERNAME \
  --language python \
  --after 2024-01-01 \
  --max-files 10
```

### Import Options

| Option | Description |
|--------|-------------|
| `--include-private` | Import private gists |
| `--include-starred` | Import starred gists |
| `--include-forks` | Import forked gists |
| `--preserve-dates` | Keep original timestamps |
| `--download-comments` | Import gist comments |
| `--auto-tags` | Generate tags from content |
| `--skip-existing` | Don't reimport existing |
| `--update-existing` | Update if already exists |

### Handling Rate Limits

GitHub API has rate limits:
- **Authenticated**: 5,000 requests/hour
- **Per Gist**: ~3-5 API calls

```bash
# Check rate limit
./casgists import github-rate-limit --token YOUR_TOKEN

# Import with rate limit handling
./casgists import github \
  --token YOUR_TOKEN \
  --respect-rate-limit \
  --delay 2s
```

### Post-Import Tasks

```bash
# Verify import
./casgists import verify \
  --source github \
  --report import-report.json

# Fix permissions
./casgists gists fix-permissions \
  --owner YOUR_USERNAME

# Rebuild search index
./casgists search reindex
```

## Database Migrations

### Running Migrations

```bash
# Check pending migrations
./casgists db migrate status

# Run all migrations
./casgists db migrate up

# Rollback last migration
./casgists db migrate down

# Rollback to specific version
./casgists db migrate to VERSION
```

### Creating Custom Migrations

```bash
# Generate migration file
./casgists db migrate create add_custom_field

# Edit migration file
# migrations/TIMESTAMP_add_custom_field.sql
```

Example migration:
```sql
-- +migrate Up
ALTER TABLE gists ADD COLUMN custom_field VARCHAR(255);
CREATE INDEX idx_gists_custom_field ON gists(custom_field);

-- +migrate Down
DROP INDEX idx_gists_custom_field;
ALTER TABLE gists DROP COLUMN custom_field;
```

### Migration Best Practices

1. **Always Backup First**
   ```bash
   ./casgists db backup --output pre-migration.sql
   ```

2. **Test on Staging**
   ```bash
   # Copy production data
   pg_dump prod_db | psql staging_db
   
   # Test migration
   ./casgists db migrate up --dry-run
   ```

3. **Monitor Migration**
   ```bash
   # Watch progress
   ./casgists db migrate up --verbose --progress
   ```

## Version Upgrades

### Upgrade Process

#### 1. Check Compatibility
```bash
# Check current version
./casgists version

# Check upgrade path
./casgists upgrade check --to v2.0.0
```

#### 2. Backup Everything
```bash
# Full backup
./casgists backup --full \
  --include-repos \
  --include-uploads \
  --output backup-v1.tar.gz

# Database only
./casgists db backup --output db-backup.sql
```

#### 3. Perform Upgrade
```bash
# Stop service
systemctl stop casgists

# Backup binary
cp casgists casgists.backup

# Download new version
curl -L https://github.com/casapps/casgists/releases/download/v2.0.0/casgists-linux-amd64 -o casgists
chmod +x casgists

# Run upgrade
./casgists upgrade --from v1.0.0 --to v2.0.0

# Start service
systemctl start casgists
```

#### 4. Verify Upgrade
```bash
# Check version
./casgists version

# Run health checks
./casgists healthcheck --full

# Verify features
./casgists test --suite upgrade
```

### Rolling Back

```bash
# Stop service
systemctl stop casgists

# Restore binary
mv casgists.backup casgists

# Restore database
./casgists db restore --input db-backup.sql

# Start service
systemctl start casgists
```

## Data Export/Import

### Full Export

```bash
# Export everything
./casgists export \
  --output export.tar.gz \
  --include users,gists,organizations \
  --format json
```

### Selective Export

```bash
# Export specific user data
./casgists export user \
  --username john \
  --include gists,stars,comments \
  --output john-export.json

# Export date range
./casgists export gists \
  --after 2024-01-01 \
  --before 2024-12-31 \
  --output 2024-gists.json

# Export by language
./casgists export gists \
  --language python,javascript \
  --output code-gists.json
```

### Import Data

```bash
# Import full export
./casgists import \
  --input export.tar.gz \
  --skip-existing

# Import with mapping
./casgists import \
  --input export.json \
  --user-mapping users.map \
  --update-existing

# Dry run
./casgists import \
  --input export.json \
  --dry-run \
  --verbose
```

### Export Formats

#### JSON Format
```json
{
  "version": "1.0",
  "exported_at": "2024-01-01T00:00:00Z",
  "users": [...],
  "gists": [...],
  "organizations": [...]
}
```

#### CSV Format
```csv
type,id,username,email,created_at
user,123,john,john@example.com,2024-01-01
```

## Migration Checklist

### Pre-Migration
- [ ] Backup source data
- [ ] Test migration process
- [ ] Notify users of downtime
- [ ] Prepare rollback plan
- [ ] Verify storage space

### During Migration
- [ ] Monitor progress
- [ ] Check error logs
- [ ] Verify data integrity
- [ ] Test sample data

### Post-Migration
- [ ] Run verification tests
- [ ] Update DNS/redirects
- [ ] Send user notifications
- [ ] Monitor performance
- [ ] Keep backups for 30 days

## Common Issues

### Memory Issues
```bash
# Increase memory for large migrations
export CASGISTS_MIGRATION_MEMORY=4G
./casgists migrate ...
```

### Timeout Issues
```bash
# Increase timeouts
./casgists migrate \
  --connection-timeout 60s \
  --query-timeout 300s
```

### Character Encoding
```bash
# Force UTF-8
export LC_ALL=en_US.UTF-8
./casgists migrate \
  --encoding utf8
```

## Support

For migration assistance:
- Documentation: [docs.casgists.com/migration](https://docs.casgists.com/migration)
- Issues: [github.com/casapps/casgists/issues](https://github.com/casapps/casgists/issues)
- Community: [forum.casgists.com/migration](https://forum.casgists.com/migration)