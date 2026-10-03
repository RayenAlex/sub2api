-- Store bounded, display-only web_search_call details under the parent usage log.
-- Child rows never participate in usage aggregation, billing, or pagination.
CREATE TABLE IF NOT EXISTS usage_web_search_events (
    id BIGSERIAL PRIMARY KEY,
    usage_log_id BIGINT NOT NULL REFERENCES usage_logs(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    call_id TEXT NOT NULL DEFAULT '',
    query TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    source_count INTEGER NOT NULL DEFAULT 0 CHECK (source_count >= 0),
    sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT usage_web_search_events_usage_log_sequence_key UNIQUE (usage_log_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_usage_web_search_events_call_id
    ON usage_web_search_events (call_id);
