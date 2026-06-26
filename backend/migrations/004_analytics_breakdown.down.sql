DROP TABLE IF EXISTS clicks_by_referer;
DROP TABLE IF EXISTS clicks_by_browser;
DROP TABLE IF EXISTS clicks_by_device;

ALTER TABLE clicks ADD COLUMN IF NOT EXISTS user_agent TEXT;
ALTER TABLE clicks ADD COLUMN IF NOT EXISTS referer    TEXT;
