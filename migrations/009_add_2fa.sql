-- Migration: Add 2FA support to users table
-- Description: Add TOTP secret, backup codes, and 2FA status fields

-- Add 2FA columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS backup_codes JSONB DEFAULT '[]'::jsonb;

-- Create index for 2FA enabled users
CREATE INDEX IF NOT EXISTS idx_users_totp_enabled ON users(totp_enabled) WHERE totp_enabled = TRUE;

-- Add column for tracking last 2FA verification (session management)
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_totp_verified TIMESTAMP WITH TIME ZONE;
