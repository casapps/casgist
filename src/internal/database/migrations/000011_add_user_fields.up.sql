-- Add missing user fields for testing compatibility

-- Add is_suspended column
ALTER TABLE users ADD COLUMN is_suspended BOOLEAN DEFAULT FALSE;

-- Add is_email_verified column (to match the model)
ALTER TABLE users ADD COLUMN is_email_verified BOOLEAN DEFAULT FALSE;

-- Update existing email_verified values to is_email_verified
UPDATE users SET is_email_verified = email_verified WHERE email_verified IS NOT NULL;