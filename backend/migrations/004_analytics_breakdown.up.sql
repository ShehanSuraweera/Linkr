-- Parse-at-collection-time analytics: device, browser, and referrer domain are
-- derived from user_agent/referer in the pipeline and stored as pre-aggregated
-- counts. Raw personal signals are never persisted beyond the transaction.

ALTER TABLE clicks DROP COLUMN IF EXISTS user_agent;
ALTER TABLE clicks DROP COLUMN IF EXISTS referer;

CREATE TABLE clicks_by_device (
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    day     DATE   NOT NULL,
    device  TEXT   NOT NULL, -- 'desktop' | 'mobile' | 'tablet'
    count   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (link_id, day, device)
);

CREATE TABLE clicks_by_browser (
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    day     DATE   NOT NULL,
    browser TEXT   NOT NULL, -- 'chrome' | 'firefox' | 'safari' | 'edge' | 'opera' | 'other'
    count   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (link_id, day, browser)
);

CREATE TABLE clicks_by_referer (
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    day     DATE   NOT NULL,
    domain  TEXT   NOT NULL, -- hostname or 'direct'
    count   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (link_id, day, domain)
);

CREATE INDEX idx_clicks_by_device_link  ON clicks_by_device  (link_id);
CREATE INDEX idx_clicks_by_browser_link ON clicks_by_browser (link_id);
CREATE INDEX idx_clicks_by_referer_link ON clicks_by_referer (link_id);
