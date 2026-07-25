-- Push subscriptions (the rest of auth is in 001_users_sessions.sql)
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id        TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint  TEXT NOT NULL UNIQUE,
    auth      TEXT NOT NULL,
    p256dh    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user ON push_subscriptions(user_id);
