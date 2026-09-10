-- cache_diagnostic contains opt-in, HMAC-only OpenCode cache diagnostics.
-- It deliberately never stores raw prompts, tool arguments, session IDs,
-- credentials, cookies, or model output. NULL means diagnostics were disabled
-- for the selected account or the request did not use the raw Chat Completions path.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS cache_diagnostic JSONB NULL;

COMMENT ON COLUMN usage_logs.cache_diagnostic IS
    'Opt-in non-reversible OpenCode cache diagnostics: request/session/tool HMACs plus message roles and count';
