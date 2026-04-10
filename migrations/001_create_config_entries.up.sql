CREATE TABLE IF NOT EXISTS config_entries (
    id         BIGSERIAL PRIMARY KEY,
    namespace  TEXT NOT NULL,
    key        TEXT NOT NULL,
    value      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (namespace, key)
);

CREATE INDEX idx_config_entries_namespace ON config_entries (namespace);
