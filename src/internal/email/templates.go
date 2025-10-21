package email

import (
	"bytes"
	"fmt"
	"html/template"
	textTemplate "text/template"
)

// TemplateRenderer handles email template rendering
type TemplateRenderer struct {
	htmlTemplates map[EmailType]*template.Template
	textTemplates map[EmailType]*textTemplate.Template
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		htmlTemplates: make(map[EmailType]*template.Template),
		textTemplates: make(map[EmailType]*textTemplate.Template),
	}
}

// LoadDefaultTemplates loads built-in email templates
func (tr *TemplateRenderer) LoadDefaultTemplates() error {
	// Load all default templates
	for emailType, templates := range defaultTemplates {
		// Load HTML template
		if templates.HTML != "" {
			htmlTmpl, err := template.New(string(emailType)).Parse(templates.HTML)
			if err != nil {
				return fmt.Errorf("failed to parse HTML template for %s: %w", emailType, err)
			}
			tr.htmlTemplates[emailType] = htmlTmpl
		}

		// Load text template
		if templates.Text != "" {
			textTmpl, err := textTemplate.New(string(emailType)).Parse(templates.Text)
			if err != nil {
				return fmt.Errorf("failed to parse text template for %s: %w", emailType, err)
			}
			tr.textTemplates[emailType] = textTmpl
		}
	}

	return nil
}

// RenderHTML renders an email template to HTML
func (tr *TemplateRenderer) RenderHTML(emailType EmailType, data EmailData) (string, error) {
	tmpl, exists := tr.htmlTemplates[emailType]
	if !exists {
		return "", fmt.Errorf("HTML template not found for email type: %s", emailType)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}

// RenderText renders an email template to plain text
func (tr *TemplateRenderer) RenderText(emailType EmailType, data EmailData) (string, error) {
	tmpl, exists := tr.textTemplates[emailType]
	if !exists {
		return "", fmt.Errorf("text template not found for email type: %s", emailType)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}

// Template represents email template content
type Template struct {
	HTML string
	Text string
}

// Default email templates
var defaultTemplates = map[EmailType]Template{
	EmailTypeVerification: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Verify Your Email - CasGists</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">Verify Your Email Address</h1>
        <p>Hello {{.UserName}},</p>
        <p>Welcome to CasGists! Please verify your email address by clicking the button below:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationURL}}" style="background-color: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Verify Email Address</a>
        </div>
        <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
        <p style="word-break: break-all; background-color: #f8f9fa; padding: 10px; border-radius: 4px;">{{.VerificationURL}}</p>
        <p><strong>This verification link will expire on {{.ExpiresAt}}.</strong></p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #666; font-size: 14px;">If you didn't create an account with CasGists, you can safely ignore this email.</p>
    </div>
</body>
</html>`,
		Text: `Verify Your Email Address - CasGists

Hello {{.UserName}},

Welcome to CasGists! Please verify your email address by visiting this link:

{{.VerificationURL}}

This verification link will expire on {{.ExpiresAt}}.

If you didn't create an account with CasGists, you can safely ignore this email.`,
	},

	EmailTypePasswordReset: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Reset Your Password - CasGists</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #e74c3c;">Reset Your Password</h1>
        <p>Hello {{.UserName}},</p>
        <p>We received a request to reset your password for your CasGists account. Click the button below to reset it:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background-color: #e74c3c; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
        </div>
        <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
        <p style="word-break: break-all; background-color: #f8f9fa; padding: 10px; border-radius: 4px;">{{.ResetURL}}</p>
        <p><strong>This reset link will expire on {{.ExpiresAt}}.</strong></p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #666; font-size: 14px;">If you didn't request a password reset, you can safely ignore this email. Your password will not be changed.</p>
    </div>
</body>
</html>`,
		Text: `Reset Your Password - CasGists

Hello {{.UserName}},

We received a request to reset your password for your CasGists account. Please visit this link to reset it:

{{.ResetURL}}

This reset link will expire on {{.ExpiresAt}}.

If you didn't request a password reset, you can safely ignore this email. Your password will not be changed.`,
	},

	EmailTypeWelcome: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Welcome to CasGists!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #27ae60;">Welcome to CasGists!</h1>
        <p>Hello {{.UserName}},</p>
        <p>Welcome to CasGists - your self-hosted code snippet sharing platform! We're excited to have you on board.</p>
        <h2>Getting Started</h2>
        <ul>
            <li>Create your first gist and share your code</li>
            <li>Explore public gists from other users</li>
            <li>Star interesting gists to save them for later</li>
            <li>Follow other developers to see their latest work</li>
        </ul>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.LoginURL}}" style="background-color: #27ae60; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Get Started</a>
        </div>
        <p>If you have any questions or need help, please don't hesitate to <a href="{{.SupportURL}}">contact our support team</a>.</p>
        <p>Happy coding!</p>
        <p>The CasGists Team</p>
    </div>
</body>
</html>`,
		Text: `Welcome to CasGists!

Hello {{.UserName}},

Welcome to CasGists - your self-hosted code snippet sharing platform! We're excited to have you on board.

Getting Started:
- Create your first gist and share your code
- Explore public gists from other users
- Star interesting gists to save them for later
- Follow other developers to see their latest work

Get started: {{.LoginURL}}

If you have any questions or need help, please contact our support team: {{.SupportURL}}

Happy coding!
The CasGists Team`,
	},

	EmailTypeGistStarred: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Your gist was starred!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #f39c12;">⭐ Your gist was starred!</h1>
        <p>Hello {{.RecipientName}},</p>
        <p><strong>{{.ActorName}}</strong> just starred your gist "<strong>{{.GistTitle}}</strong>"!</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.GistURL}}" style="background-color: #f39c12; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">View Gist</a>
        </div>
        <p>Keep up the great work!</p>
    </div>
</body>
</html>`,
		Text: `⭐ Your gist was starred!

Hello {{.RecipientName}},

{{.ActorName}} just starred your gist "{{.GistTitle}}"!

View your gist: {{.GistURL}}

Keep up the great work!`,
	},

	EmailTypeUserFollowed: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>You have a new follower!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #9b59b6;">👥 You have a new follower!</h1>
        <p>Hello {{.RecipientName}},</p>
        <p><strong>{{.FollowerName}}</strong> is now following you on CasGists!</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.FollowerURL}}" style="background-color: #9b59b6; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">View Profile</a>
        </div>
        <p>Check out <a href="{{.ProfileURL}}">your profile</a> to see all your followers.</p>
    </div>
</body>
</html>`,
		Text: `👥 You have a new follower!

Hello {{.RecipientName}},

{{.FollowerName}} is now following you on CasGists!

View their profile: {{.FollowerURL}}
View your profile: {{.ProfileURL}}`,
	},

	EmailTypeGistForked: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Your gist was forked!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #27ae60;">🍴 Your gist was forked!</h1>
        <p>Hello {{.RecipientName}},</p>
        <p><strong>{{.ActorName}}</strong> just forked your gist "<strong>{{.GistTitle}}</strong>"!</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ForkedGistURL}}" style="background-color: #27ae60; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">View Fork</a>
        </div>
        <p>View the original gist: <a href="{{.GistURL}}">{{.GistTitle}}</a></p>
    </div>
</body>
</html>`,
		Text: `🍴 Your gist was forked!

Hello {{.RecipientName}},

{{.ActorName}} just forked your gist "{{.GistTitle}}"!

View the fork: {{.ForkedGistURL}}
View the original: {{.GistURL}}`,
	},

	EmailTypeWeeklyDigest: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Your weekly CasGists digest</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #3498db;">📊 Your Weekly CasGists Digest</h1>
        <p>Hello {{.UserName}},</p>
        <p>Here's what happened this week:</p>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px; margin: 20px 0;">
            <h3 style="color: #2c3e50; margin-top: 0;">📈 Your Stats</h3>
            <ul style="list-style: none; padding: 0;">
                <li>📝 <strong>{{.TotalGists}}</strong> total gists ({{.NewGists}} new this week)</li>
                <li>⭐ <strong>{{.TotalStars}}</strong> total stars ({{.NewStars}} new this week)</li>
                <li>🍴 <strong>{{.TotalForks}}</strong> total forks ({{.NewForks}} new this week)</li>
                <li>👥 <strong>{{.TotalFollowers}}</strong> followers ({{.NewFollowers}} new this week)</li>
            </ul>
        </div>

        {{if .PopularGists}}
        <h3 style="color: #2c3e50;">🔥 Your Most Popular Gists</h3>
        <ol>
            {{range .PopularGists}}
            <li><a href="{{.URL}}">{{.Title}}</a> - {{.Stars}} stars, {{.Forks}} forks</li>
            {{end}}
        </ol>
        {{end}}

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.DashboardURL}}" style="background-color: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">View Dashboard</a>
        </div>

        <p style="color: #666; font-size: 14px;">You can adjust your email preferences in your <a href="{{.SettingsURL}}">account settings</a>.</p>
    </div>
</body>
</html>`,
		Text: `📊 Your Weekly CasGists Digest

Hello {{.UserName}},

Here's what happened this week:

📈 Your Stats
- 📝 {{.TotalGists}} total gists ({{.NewGists}} new this week)
- ⭐ {{.TotalStars}} total stars ({{.NewStars}} new this week)  
- 🍴 {{.TotalForks}} total forks ({{.NewForks}} new this week)
- 👥 {{.TotalFollowers}} followers ({{.NewFollowers}} new this week)

View your dashboard: {{.DashboardURL}}

You can adjust your email preferences in your account settings: {{.SettingsURL}}`,
	},

	EmailTypeSystemAlert: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>CasGists System Alert</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #e74c3c;">🚨 System Alert</h1>
        <p>Hello {{.UserName}},</p>
        
        <div style="background-color: #fff5f5; border: 1px solid #e74c3c; padding: 20px; border-radius: 8px; margin: 20px 0;">
            <h3 style="color: #e74c3c; margin-top: 0;">{{.AlertTitle}}</h3>
            <p>{{.AlertMessage}}</p>
            {{if .ActionRequired}}
            <p><strong>Action Required:</strong> {{.ActionDescription}}</p>
            {{end}}
        </div>

        {{if .ActionURL}}
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ActionURL}}" style="background-color: #e74c3c; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">{{.ActionLabel}}</a>
        </div>
        {{end}}

        <p style="color: #666; font-size: 14px;">If you have questions, please contact support: {{.SupportEmail}}</p>
    </div>
</body>
</html>`,
		Text: `🚨 System Alert

Hello {{.UserName}},

{{.AlertTitle}}

{{.AlertMessage}}

{{if .ActionRequired}}Action Required: {{.ActionDescription}}{{end}}

{{if .ActionURL}}{{.ActionLabel}}: {{.ActionURL}}{{end}}

If you have questions, please contact support: {{.SupportEmail}}`,
	},

	EmailTypeInvitation: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>You're invited to join CasGists!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #3498db;">🎉 You're Invited!</h1>
        <p>Hello,</p>
        <p><strong>{{.InviterName}}</strong> has invited you to join {{if .OrganizationName}}the <strong>{{.OrganizationName}}</strong> organization on {{end}}CasGists!</p>
        
        {{if .PersonalMessage}}
        <div style="background-color: #f8f9fa; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <p style="font-style: italic; margin: 0;">"{{.PersonalMessage}}"</p>
            <p style="text-align: right; color: #666; margin: 5px 0 0 0;">- {{.InviterName}}</p>
        </div>
        {{end}}

        <p>CasGists is a secure, self-hosted code snippet and documentation platform that helps teams collaborate efficiently.</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.InviteURL}}" style="background-color: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Accept Invitation</a>
        </div>

        <p><strong>This invitation will expire on {{.ExpiresAt}}.</strong></p>

        <p style="color: #666; font-size: 14px;">If you don't want to accept this invitation, you can safely ignore this email.</p>
    </div>
</body>
</html>`,
		Text: `🎉 You're Invited!

Hello,

{{.InviterName}} has invited you to join {{if .OrganizationName}}the {{.OrganizationName}} organization on {{end}}CasGists!

{{if .PersonalMessage}}"{{.PersonalMessage}}" - {{.InviterName}}{{end}}

CasGists is a secure, self-hosted code snippet and documentation platform that helps teams collaborate efficiently.

Accept the invitation: {{.InviteURL}}

This invitation will expire on {{.ExpiresAt}}.

If you don't want to accept this invitation, you can safely ignore this email.`,
	},

	EmailTypeGistCommented: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>New comment on your gist</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #3498db;">💬 New comment on your gist!</h1>
        <p>Hello {{.RecipientName}},</p>
        <p><strong>{{.CommenterName}}</strong> just commented on your gist "<strong>{{.GistTitle}}</strong>":</p>
        
        <div style="background-color: #f8f9fa; padding: 15px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #3498db;">
            <p style="margin: 0; white-space: pre-wrap;">{{.CommentPreview}}</p>
        </div>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.GistURL}}#comment-{{.CommentID}}" style="background-color: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">View Comment</a>
        </div>
        
        <p style="color: #666; font-size: 14px;">You can adjust your notification settings in your <a href="{{.SettingsURL}}">account settings</a>.</p>
    </div>
</body>
</html>`,
		Text: `💬 New comment on your gist!

Hello {{.RecipientName}},

{{.CommenterName}} just commented on your gist "{{.GistTitle}}":

{{.CommentPreview}}

View the full comment: {{.GistURL}}#comment-{{.CommentID}}

You can adjust your notification settings in your account settings: {{.SettingsURL}}`,
	},

	EmailTypeBackupComplete: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Backup completed successfully</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #27ae60;">✅ Backup Completed Successfully!</h1>
        <p>Hello {{.UserName}},</p>
        <p>Your CasGists backup has been completed successfully.</p>
        
        <div style="background-color: #d4edda; padding: 20px; border-radius: 8px; margin: 20px 0; border: 1px solid #c3e6cb;">
            <h3 style="color: #155724; margin-top: 0;">Backup Details:</h3>
            <ul style="list-style: none; padding: 0;">
                <li>📅 <strong>Date:</strong> {{.BackupDate}}</li>
                <li>📊 <strong>Size:</strong> {{.BackupSize}}</li>
                <li>📝 <strong>Gists backed up:</strong> {{.GistCount}}</li>
                <li>👥 <strong>Users backed up:</strong> {{.UserCount}}</li>
                <li>💾 <strong>Storage location:</strong> {{.StorageLocation}}</li>
                {{if .BackupType}}<li>🔧 <strong>Type:</strong> {{.BackupType}}</li>{{end}}
            </ul>
        </div>
        
        {{if .DownloadURL}}
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.DownloadURL}}" style="background-color: #27ae60; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Download Backup</a>
        </div>
        {{end}}
        
        <p><strong>Next scheduled backup:</strong> {{.NextBackupDate}}</p>
        
        <p style="color: #666; font-size: 14px;">Backup settings can be adjusted in your <a href="{{.AdminURL}}">admin panel</a>.</p>
    </div>
</body>
</html>`,
		Text: `✅ Backup Completed Successfully!

Hello {{.UserName}},

Your CasGists backup has been completed successfully.

Backup Details:
- 📅 Date: {{.BackupDate}}
- 📊 Size: {{.BackupSize}}
- 📝 Gists backed up: {{.GistCount}}
- 👥 Users backed up: {{.UserCount}}
- 💾 Storage location: {{.StorageLocation}}
{{if .BackupType}}- 🔧 Type: {{.BackupType}}{{end}}

{{if .DownloadURL}}Download backup: {{.DownloadURL}}{{end}}

Next scheduled backup: {{.NextBackupDate}}

Backup settings can be adjusted in your admin panel: {{.AdminURL}}`,
	},

	EmailTypeMigrationComplete: {
		HTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Migration completed successfully</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #27ae60;">🎉 Migration Completed Successfully!</h1>
        <p>Hello {{.UserName}},</p>
        <p>Your migration to CasGists has been completed successfully. All your data has been imported and is ready to use.</p>
        
        <div style="background-color: #d4edda; padding: 20px; border-radius: 8px; margin: 20px 0; border: 1px solid #c3e6cb;">
            <h3 style="color: #155724; margin-top: 0;">Migration Summary:</h3>
            <ul style="list-style: none; padding: 0;">
                <li>📅 <strong>Migration date:</strong> {{.MigrationDate}}</li>
                <li>🔄 <strong>Source:</strong> {{.SourcePlatform}}</li>
                <li>⏱️ <strong>Duration:</strong> {{.Duration}}</li>
                <li>📝 <strong>Gists imported:</strong> {{.GistCount}}</li>
                <li>⭐ <strong>Stars imported:</strong> {{.StarCount}}</li>
                <li>👥 <strong>Followers imported:</strong> {{.FollowerCount}}</li>
                {{if .SkippedItems}}<li>⚠️ <strong>Items skipped:</strong> {{.SkippedItems}}</li>{{end}}
            </ul>
        </div>

        {{if .SkippedItems}}
        <div style="background-color: #fff3cd; padding: 15px; border-radius: 8px; margin: 20px 0; border: 1px solid #ffeeba;">
            <p style="margin: 0;"><strong>Note:</strong> Some items could not be migrated. Please check the <a href="{{.MigrationReportURL}}">migration report</a> for details.</p>
        </div>
        {{end}}
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.DashboardURL}}" style="background-color: #27ae60; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Go to Dashboard</a>
        </div>
        
        <h3 style="color: #2c3e50;">What's Next?</h3>
        <ul>
            <li>Review your imported gists to ensure everything looks correct</li>
            <li>Update your profile and preferences</li>
            <li>Explore CasGists features like organizations and advanced search</li>
            <li>Set up integrations and API access if needed</li>
        </ul>
        
        <p style="color: #666; font-size: 14px;">If you encounter any issues, please contact <a href="{{.SupportURL}}">support</a>.</p>
    </div>
</body>
</html>`,
		Text: `🎉 Migration Completed Successfully!

Hello {{.UserName}},

Your migration to CasGists has been completed successfully. All your data has been imported and is ready to use.

Migration Summary:
- 📅 Migration date: {{.MigrationDate}}
- 🔄 Source: {{.SourcePlatform}}
- ⏱️ Duration: {{.Duration}}
- 📝 Gists imported: {{.GistCount}}
- ⭐ Stars imported: {{.StarCount}}
- 👥 Followers imported: {{.FollowerCount}}
{{if .SkippedItems}}- ⚠️ Items skipped: {{.SkippedItems}}{{end}}

{{if .SkippedItems}}Note: Some items could not be migrated. Please check the migration report: {{.MigrationReportURL}}{{end}}

Go to your dashboard: {{.DashboardURL}}

What's Next?
- Review your imported gists to ensure everything looks correct
- Update your profile and preferences
- Explore CasGists features like organizations and advanced search
- Set up integrations and API access if needed

If you encounter any issues, please contact support: {{.SupportURL}}`,
	},
}

// GetDefaultSubjects returns default email subjects
func GetDefaultSubjects() map[EmailType]string {
	return map[EmailType]string{
		EmailTypeVerification:  "Verify your CasGists account",
		EmailTypePasswordReset: "Reset your CasGists password",
		EmailTypeWelcome:       "Welcome to CasGists!",
		EmailTypeGistStarred:   "⭐ Your gist was starred!",
		EmailTypeGistForked:    "🍴 Your gist was forked!",
		EmailTypeUserFollowed:  "👥 You have a new follower!",
		EmailTypeWeeklyDigest:  "📊 Your weekly CasGists digest",
		EmailTypeSystemAlert:   "🚨 CasGists system alert",
		EmailTypeInvitation:       "You're invited to join CasGists",
		EmailTypeGistCommented:    "💬 New comment on your gist",
		EmailTypeBackupComplete:   "✅ Backup completed successfully",
		EmailTypeMigrationComplete: "🎉 Migration completed successfully",
	}
}

// RenderSubject renders email subject with template variables
func RenderSubject(subject string, data EmailData) (string, error) {
	tmpl, err := textTemplate.New("subject").Parse(subject)
	if err != nil {
		return subject, nil // Return original subject if parsing fails
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return subject, nil // Return original subject if execution fails
	}

	return buf.String(), nil
}