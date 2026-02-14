-- MCPedia schema
-- Set these PRAGMAs at connection time, not here:
--   PRAGMA journal_mode=WAL;
--   PRAGMA foreign_keys=ON;
--   PRAGMA busy_timeout=5000;

CREATE TABLE IF NOT EXISTS entries (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT UNIQUE NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    content     TEXT NOT NULL CHECK(length(content) <= 32768),
    kind        TEXT NOT NULL DEFAULT 'skill',
    language    TEXT NOT NULL DEFAULT '',
    domain      TEXT NOT NULL DEFAULT '',
    project     TEXT NOT NULL DEFAULT '',
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tags (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS entry_tags (
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    tag_id   INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, tag_id)
);

-- Usage statistics
CREATE TABLE IF NOT EXISTS entry_stats (
    entry_id       INTEGER PRIMARY KEY REFERENCES entries(id) ON DELETE CASCADE,
    reads          INTEGER NOT NULL DEFAULT 0,
    searches       INTEGER NOT NULL DEFAULT 0,
    updates        INTEGER NOT NULL DEFAULT 0,
    last_read_at   TEXT,
    last_search_at TEXT,
    last_update_at TEXT
);

-- Write-lock (single row enforced)
CREATE TABLE IF NOT EXISTS lock (
    id     INTEGER PRIMARY KEY CHECK (id = 1),
    active INTEGER NOT NULL DEFAULT 0,
    token  TEXT NOT NULL DEFAULT ''
);

INSERT OR IGNORE INTO lock (id, active, token) VALUES (1, 0, '');

-- Indexes
CREATE INDEX IF NOT EXISTS idx_entries_language ON entries(language);
CREATE INDEX IF NOT EXISTS idx_entries_domain   ON entries(domain);
CREATE INDEX IF NOT EXISTS idx_entries_kind     ON entries(kind);
CREATE INDEX IF NOT EXISTS idx_entries_project  ON entries(project);
CREATE INDEX IF NOT EXISTS idx_tags_name        ON tags(name);

-- FTS5 virtual table for full-text search
-- Note: FTS5 does not support IF NOT EXISTS, so we handle this in Go code
-- by checking if the table already exists before creating it.
