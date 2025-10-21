# Email Templates

This directory contains customizable email templates for CasGists. These templates override the default built-in templates.

## Template Files

Email templates are organized by type:
- `verification/` - Email verification templates
- `password-reset/` - Password reset templates
- `welcome/` - Welcome email templates
- `notifications/` - Various notification templates (stars, forks, comments, follows)
- `system/` - System alert and backup notification templates
- `digest/` - Weekly digest templates

## Template Format

Each email type should have two files:
- `template.html` - HTML version of the email
- `template.txt` - Plain text version of the email
- `subject.txt` - Email subject line (optional, uses default if not provided)

## Template Variables

Templates use Go's template syntax. Available variables vary by email type:

### Common Variables
- `{{.UserName}}` - Recipient's username
- `{{.ServerURL}}` - Base server URL

### Verification Email
- `{{.VerificationURL}}` - Email verification link
- `{{.ExpiresAt}}` - Expiration time

### Password Reset
- `{{.ResetURL}}` - Password reset link
- `{{.ExpiresAt}}` - Expiration time

### Comment Notification
- `{{.RecipientName}}` - Recipient's name
- `{{.CommenterName}}` - Name of person who commented
- `{{.GistTitle}}` - Title of the gist
- `{{.CommentPreview}}` - Preview of the comment
- `{{.GistURL}}` - Link to the gist
- `{{.CommentID}}` - Comment ID for anchoring

### Backup Complete
- `{{.BackupDate}}` - Date of backup
- `{{.BackupSize}}` - Size of backup
- `{{.GistCount}}` - Number of gists backed up
- `{{.UserCount}}` - Number of users backed up
- `{{.StorageLocation}}` - Where backup is stored
- `{{.DownloadURL}}` - Link to download backup (optional)
- `{{.NextBackupDate}}` - Next scheduled backup

### Migration Complete
- `{{.MigrationDate}}` - Date of migration
- `{{.SourcePlatform}}` - Source platform (GitHub, GitLab, etc.)
- `{{.Duration}}` - How long migration took
- `{{.GistCount}}` - Number of gists migrated
- `{{.StarCount}}` - Number of stars migrated
- `{{.FollowerCount}}` - Number of followers migrated
- `{{.SkippedItems}}` - Number of items skipped (if any)
- `{{.MigrationReportURL}}` - Link to detailed report

## Customization

To customize an email template:
1. Create the appropriate directory structure
2. Copy the template content from the built-in templates
3. Modify as needed
4. Restart CasGists to load the new templates

Templates are loaded at startup. Changes require a restart to take effect.