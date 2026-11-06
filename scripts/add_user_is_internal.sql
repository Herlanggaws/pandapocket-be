-- Optional explicit DDL for users.is_internal (also applied via GORM AutoMigrate).
-- Safe to re-run. Do NOT run against prod without review.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT false;
