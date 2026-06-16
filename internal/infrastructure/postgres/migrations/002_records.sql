CREATE TABLE IF NOT EXISTS keeper_records (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES keeper_users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    meta TEXT NOT NULL DEFAULT '',
    ciphertext BYTEA NOT NULL DEFAULT '',
    version BIGINT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS keeper_records_owner_version_idx ON keeper_records(owner_id, version);
