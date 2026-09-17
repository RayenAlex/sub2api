-- Ensure token quota CHECK constraints exist for databases that reached migration 236
-- before the constraints were recorded. This is intentionally a new migration: migration
-- 236 is immutable after deployment and its checksum must continue matching schema_migrations.
DO $$ BEGIN
    ALTER TABLE groups ADD CONSTRAINT chk_groups_daily_token_limit_nonneg CHECK (daily_token_limit IS NULL OR daily_token_limit >= 0);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE groups ADD CONSTRAINT chk_groups_weekly_token_limit_nonneg CHECK (weekly_token_limit IS NULL OR weekly_token_limit >= 0);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE groups ADD CONSTRAINT chk_groups_monthly_token_limit_nonneg CHECK (monthly_token_limit IS NULL OR monthly_token_limit >= 0);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE user_subscriptions ADD CONSTRAINT chk_user_subscriptions_daily_token_usage_nonneg CHECK (daily_token_usage >= 0);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE user_subscriptions ADD CONSTRAINT chk_user_subscriptions_weekly_token_usage_nonneg CHECK (weekly_token_usage >= 0);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE user_subscriptions ADD CONSTRAINT chk_user_subscriptions_monthly_token_usage_nonneg CHECK (monthly_token_usage >= 0);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
