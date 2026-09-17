-- Token 配额订阅分组：每日/每周/每月 token 上限 + 用量
-- groups: 限额（NULL=不限，0=禁止消耗）
ALTER TABLE groups ADD COLUMN IF NOT EXISTS daily_token_limit BIGINT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS weekly_token_limit BIGINT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS monthly_token_limit BIGINT;

-- user_subscriptions: 用量计数（默认 0）
ALTER TABLE user_subscriptions ADD COLUMN IF NOT EXISTS daily_token_usage BIGINT NOT NULL DEFAULT 0;
ALTER TABLE user_subscriptions ADD COLUMN IF NOT EXISTS weekly_token_usage BIGINT NOT NULL DEFAULT 0;
ALTER TABLE user_subscriptions ADD COLUMN IF NOT EXISTS monthly_token_usage BIGINT NOT NULL DEFAULT 0;

-- 约束：限额非负（迁移重试时约束可能已由前一次部分执行创建）
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

-- 约束：用量非负
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
