-- Rollback: Anonymous User Support
-- WARNING: This will fail if any anonymous users exist in the database

-- Drop cleanup audit table
DROP TABLE IF EXISTS anonymous_user_cleanups;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_user_type;
DROP INDEX IF EXISTS idx_users_anonymous_expires_at;

-- Drop constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_type_expiry_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_type_credentials_check;

-- Restore NOT NULL constraints (will fail if anonymous users exist)
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE users ALTER COLUMN name SET NOT NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Drop new columns
ALTER TABLE users DROP COLUMN IF EXISTS expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS user_type;

-- Drop enum type
DROP TYPE IF EXISTS user_type;
