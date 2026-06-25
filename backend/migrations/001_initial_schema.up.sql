CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE links (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    short_code   TEXT        NOT NULL UNIQUE,
    original_url TEXT        NOT NULL,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ,
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    deleted_at   TIMESTAMPTZ
);

CREATE TABLE clicks (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    link_id    BIGINT      NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ip_hash    TEXT,
    user_agent TEXT,
    referer    TEXT
);

CREATE TABLE click_daily (
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    day     DATE   NOT NULL,
    count   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (link_id, day)
);

-- Redirect hot path: unique index on short_code (auto-created by UNIQUE above).
-- List query: partial index for live rows, keyset pagination by (user_id, created_at DESC, id DESC).
CREATE INDEX idx_links_user_created ON links (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

-- Stats: clicks by link over time.
CREATE INDEX idx_clicks_link_time ON clicks (link_id, clicked_at);
