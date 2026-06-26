-- idx_clicks_link_time is never used: the clicks table is write-only (CopyFrom).
-- Stats queries read from click_daily. Dropping saves write overhead on every flush.
DROP INDEX IF EXISTS idx_clicks_link_time;
