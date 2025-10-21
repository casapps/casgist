# CasGists User Guide

Welcome to CasGists! This guide will help you get started with creating, managing, and sharing your code snippets.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Creating Your First Gist](#creating-your-first-gist)
3. [Managing Gists](#managing-gists)
4. [Sharing and Collaboration](#sharing-and-collaboration)
5. [Advanced Features](#advanced-features)
6. [Keyboard Shortcuts](#keyboard-shortcuts)
7. [Tips and Tricks](#tips-and-tricks)

## Getting Started

### Creating an Account

1. Navigate to your CasGists instance (e.g., `https://gists.example.com`)
2. Click **"Sign Up"** in the top right corner
3. Fill in your details:
   - **Username**: Choose a unique username (3-20 characters)
   - **Email**: Your email address for notifications
   - **Password**: Minimum 8 characters with mixed case and numbers
4. Click **"Create Account"**
5. Check your email for a verification link (if enabled)

### Logging In

1. Click **"Login"** in the top right corner
2. Enter your username/email and password
3. If you have 2FA enabled, enter your TOTP code
4. Click **"Sign In"**

### Setting Up Your Profile

1. Click your username in the top right → **"Settings"**
2. Update your profile information:
   - **Display Name**: Your full name (optional)
   - **Bio**: A short description about yourself
   - **Location**: Your location (optional)
   - **Website**: Your personal website or blog
3. Upload an avatar image (supports JPG, PNG, GIF up to 5MB)
4. Click **"Save Profile"**

## Creating Your First Gist

### Quick Create

1. Click the **"New Gist"** button in the navigation bar
2. Give your gist a title (optional but recommended)
3. Add a description to explain what your gist does
4. Choose visibility:
   - **Public**: Visible to everyone, appears in search
   - **Unlisted**: Accessible via direct link only
   - **Private**: Only visible to you
5. Add your code:
   - Enter a filename with extension (e.g., `script.py`)
   - Paste or type your code in the editor
   - The language is auto-detected from the file extension
6. Click **"Create Public/Unlisted/Private"**

### Adding Multiple Files

1. After adding your first file, click **"Add Another File"**
2. Each gist can contain up to 10 files
3. Files can be different languages and types
4. Total size limit is 10MB per gist

### File Upload

You can also drag and drop files directly:

1. Drag files from your computer to the upload area
2. Files are automatically added with their original names
3. Content is loaded into the editor for review
4. You can edit before creating the gist

### Using the Editor

The code editor supports:
- **Syntax highlighting** for 100+ languages
- **Auto-indentation**
- **Bracket matching**
- **Code folding**
- **Multiple cursors** (Ctrl/Cmd + Click)
- **Find and replace** (Ctrl/Cmd + F)

## Managing Gists

### Viewing Your Gists

1. Click **"My Gists"** in the navigation bar
2. Filter gists by:
   - **All**: Shows all your gists
   - **Public**: Only public gists
   - **Private**: Only private gists
   - **Starred**: Gists you've starred
3. Sort by:
   - **Recently created**
   - **Recently updated**
   - **Most starred**

### Editing a Gist

1. Navigate to your gist
2. Click the **"Edit"** button
3. Make your changes:
   - Update title or description
   - Modify file contents
   - Add or remove files
   - Change visibility (except public → private)
4. Click **"Update Gist"**

### Deleting a Gist

1. Navigate to your gist
2. Click the **"Edit"** button
3. Scroll to the bottom
4. Click **"Delete Gist"**
5. Confirm deletion

**Note**: Deleted gists are soft-deleted and can be recovered within 30 days by contacting support.

### Organizing with Titles and Descriptions

Best practices:
- Use descriptive titles: "Python Web Scraper for News Sites"
- Add detailed descriptions explaining:
  - What the code does
  - How to use it
  - Any dependencies required
  - Expected output
- Use hashtags in descriptions for easier searching: `#python #webscraping #beautifulsoup`

## Sharing and Collaboration

### Sharing a Gist

Each gist has multiple ways to share:

1. **Direct Link**: Copy the URL from your browser
2. **Short Link**: Click **"Copy URL"** for a shortened link
3. **Embed**: Click **"Embed"** to get HTML code for blogs/websites
4. **Raw Files**: Access raw file content via the **"Raw"** button

### Embedding Gists

To embed a gist in your blog or website:

```html
<script src="https://gists.example.com/embed/gist-id.js"></script>
```

For specific files:
```html
<script src="https://gists.example.com/embed/gist-id.js?file=specific-file.py"></script>
```

With custom styling:
```html
<div class="gist-embed" data-gist-id="gist-id" data-gist-file="file.py"></div>
<script src="https://gists.example.com/embed.js"></script>
```

### Starring Gists

- Click the **star** button to bookmark interesting gists
- View your starred gists in **"My Gists"** → **"Starred"**
- Stars help identify popular and useful gists

### Forking Gists

To create your own copy of someone else's gist:

1. Navigate to the gist
2. Click **"Fork"**
3. The forked gist appears in your account
4. Edit and customize as needed
5. The original is linked as "Forked from..."

### Comments

Engage with the community:

1. Scroll to the comments section below any gist
2. Write your comment (supports Markdown)
3. Click **"Comment"**
4. Edit or delete your own comments
5. Get notified when others comment on your gists

## Advanced Features

### Search

Find gists quickly using the search bar:

- **Basic search**: Type keywords to search titles, descriptions, and content
- **Language filter**: `language:python flask`
- **User filter**: `user:johndoe javascript`
- **Visibility filter**: `visibility:public react hooks`
- **Combined filters**: `language:go user:alice database`

Search operators:
- Quotes for exact phrases: `"hello world"`
- Exclude terms with minus: `python -django`
- Wildcard with asterisk: `print*`

### Git Access

Clone and manage gists using Git:

```bash
# Clone a gist
git clone https://gists.example.com/username/gist-id.git

# Add remote to existing repo
git remote add gist https://gists.example.com/username/gist-id.git

# Push updates
git push gist main
```

### API Access

Create an API token for programmatic access:

1. Go to **Settings** → **API Tokens**
2. Click **"Generate New Token"**
3. Give it a name and select scopes
4. Copy the token (shown only once!)
5. Use in API requests:

```bash
curl -H "Authorization: Bearer your-token" \
  https://gists.example.com/api/v1/gists
```

### Keyboard Shortcuts

#### Global Shortcuts
- `?` - Show keyboard shortcuts help
- `/` - Focus search bar
- `n` - Create new gist
- `g h` - Go to home
- `g g` - Go to my gists
- `g s` - Go to starred gists
- `g p` - Go to profile

#### Editor Shortcuts
- `Ctrl/Cmd + S` - Save gist
- `Ctrl/Cmd + Enter` - Save and close editor
- `Ctrl/Cmd + P` - Command palette
- `Ctrl/Cmd + /` - Toggle comment
- `Ctrl/Cmd + D` - Select next occurrence
- `Ctrl/Cmd + F` - Find
- `Ctrl/Cmd + H` - Find and replace
- `Alt + ↑/↓` - Move line up/down
- `Alt + Shift + ↑/↓` - Copy line up/down

#### Viewing Gists
- `e` - Edit gist (if owner)
- `f` - Fork gist
- `s` - Star/unstar gist
- `l` - Copy gist URL
- `r` - View raw file
- `.` - Focus on files

### Organizations

If enabled, you can create and join organizations:

#### Creating an Organization

1. Click your username → **"Organizations"**
2. Click **"New Organization"**
3. Choose a unique name and display name
4. Add description and details
5. Click **"Create Organization"**

#### Managing Organization Gists

- Switch context using the dropdown in navigation
- Create gists under organization ownership
- Manage member permissions:
  - **Owner**: Full control
  - **Admin**: Manage gists and members
  - **Member**: Create and edit gists

### Import/Export

#### Import from GitHub

1. Go to **Settings** → **Import**
2. Click **"Import from GitHub"**
3. Authorize with GitHub
4. Select gists to import
5. Choose visibility mapping
6. Click **"Start Import"**

#### Export Your Data

1. Go to **Settings** → **Export**
2. Choose export format:
   - **JSON**: Machine-readable format
   - **Archive**: ZIP with all gists
   - **Git Bundle**: For Git import
3. Click **"Export My Data"**

### Two-Factor Authentication (2FA)

Enhance account security:

1. Go to **Settings** → **Security**
2. Click **"Enable 2FA"**
3. Scan QR code with authenticator app:
   - Google Authenticator
   - Microsoft Authenticator
   - Authy
   - 1Password
4. Enter verification code
5. Save backup codes securely

### Email Notifications

Configure when to receive emails:

1. Go to **Settings** → **Notifications**
2. Toggle notifications for:
   - Comments on your gists
   - Stars on your gists
   - Forks of your gists
   - Weekly digest
3. Choose email frequency:
   - **Instant**: As they happen
   - **Daily**: Daily summary
   - **Weekly**: Weekly digest

## Tips and Tricks

### Productivity Tips

1. **Templates**: Save commonly used code patterns as private gists
2. **Versioning**: Use Git to track changes over time
3. **Collections**: Use consistent naming like `[Project] Description`
4. **Quick Save**: Use Ctrl+S to save without leaving the editor
5. **Markdown**: Use README.md files to document complex gists

### Formatting Code

- Use consistent indentation (spaces or tabs)
- Add comments explaining complex logic
- Include example usage in comments
- Use meaningful variable and function names
- Keep files focused on a single purpose

### Gist Ideas

- Configuration files (`.bashrc`, `.vimrc`, etc.)
- Code snippets and utilities
- Interview questions and solutions
- Learning notes and examples
- Bug reproductions
- Quick scripts and one-liners
- API examples and documentation
- SQL queries and database scripts

### Security Best Practices

1. **Never share**:
   - Passwords or API keys
   - Private SSH keys
   - Database credentials
   - Personal information
   
2. **Use environment variables** in examples:
   ```python
   api_key = os.environ.get('API_KEY')
   ```

3. **Review before sharing**: Double-check for sensitive data
4. **Use private gists** for sensitive code
5. **Rotate credentials** if accidentally exposed

### Performance Tips

- Keep gists under 1MB for best performance
- Use pagination when listing many gists
- Cache API responses when building integrations
- Compress large files before uploading
- Use raw URLs for automated downloads

## Mobile Usage

CasGists is fully responsive and works great on mobile devices:

- **Swipe gestures**: Navigate between files
- **Touch-friendly**: Large tap targets
- **Offline support**: View cached gists offline (PWA)
- **Share integration**: Share gists to other apps
- **Code viewing**: Horizontal scroll for wide code

## Troubleshooting

### Common Issues

**Can't create gists:**
- Check if you're logged in
- Verify email if required
- Check storage quota

**Syntax highlighting not working:**
- Ensure file has correct extension
- Try manually selecting language
- Clear browser cache

**Can't edit gist:**
- Verify you own the gist
- Check if you're logged in
- Ensure gist isn't locked

**Search not finding gists:**
- Wait a few minutes for indexing
- Try different search terms
- Check visibility settings

### Getting Help

- **Documentation**: https://docs.casgists.com
- **Community Forum**: https://forum.casgists.com
- **Email Support**: support@casgists.com
- **FAQ**: Check the frequently asked questions
- **Status Page**: https://status.casgists.com

## Conclusion

CasGists is designed to make sharing code simple and efficient. Whether you're storing personal snippets, sharing solutions, or collaborating with others, we hope this guide helps you make the most of CasGists.

Happy coding! 🚀