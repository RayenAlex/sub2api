-- Persist the peak multiplier used for each token-billed request so usage history can distinguish peak-priced tokens.
-- Historical and non-token usage rows stay NULL and render without a peak marker.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS applied_peak_multiplier DECIMAL(10,4);
