-- 0001_init.sql — initial GoTutor schema
-- Applied on first connect via //go:embed in embed.go.

CREATE TABLE IF NOT EXISTS chapters (
  id          TEXT PRIMARY KEY,
  title       TEXT NOT NULL,
  description TEXT NOT NULL,
  ordinal     INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS progress (
  chapter_id   TEXT PRIMARY KEY REFERENCES chapters(id),
  completed_at INTEGER,                    -- unix seconds; NULL = not completed
  attempts     INTEGER NOT NULL DEFAULT 0,
  last_output  TEXT
);

CREATE INDEX IF NOT EXISTS idx_progress_completed ON progress(completed_at);

CREATE TABLE IF NOT EXISTS settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
