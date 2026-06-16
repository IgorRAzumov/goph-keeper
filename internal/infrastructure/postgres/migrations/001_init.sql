-- Префикс keeper_ — чтобы не конфликтовать с системной/чужой таблицей users в БД postgres.
CREATE TABLE IF NOT EXISTS keeper_users (
    id UUID PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS keeper_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES keeper_users(id) ON DELETE CASCADE,
    refresh_token_hash BYTEA NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS keeper_sessions_user_id_idx ON keeper_sessions(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS keeper_sessions_refresh_token_hash_idx ON keeper_sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS keeper_sessions_refresh_expires_at_idx ON keeper_sessions(refresh_expires_at);
