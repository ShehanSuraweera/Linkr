CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_links_original_url_trgm ON links USING gin (original_url gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_short_code_trgm   ON links USING gin (short_code   gin_trgm_ops) WHERE deleted_at IS NULL;
