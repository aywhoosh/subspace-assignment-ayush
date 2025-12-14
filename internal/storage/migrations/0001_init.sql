-- +subspace Up
-- Initial schema for local-only mocknet automation PoC.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL UNIQUE,
  cookies_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  last_used_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS profiles (
  profile_id TEXT PRIMARY KEY,
  url TEXT NOT NULL UNIQUE,
  first_name TEXT,
  last_name TEXT,
  company TEXT,
  title TEXT,
  location TEXT,
  keywords TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS connection_requests (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  profile_id TEXT NOT NULL,
  sent_at TEXT NOT NULL,
  note TEXT NOT NULL,
  status TEXT NOT NULL,
  UNIQUE(profile_id),
  FOREIGN KEY(profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS connections (
  profile_id TEXT PRIMARY KEY,
  accepted_at TEXT NOT NULL,
  FOREIGN KEY(profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  thread_id TEXT NOT NULL,
  profile_id TEXT NOT NULL,
  template_id TEXT NOT NULL,
  body TEXT NOT NULL,
  sent_at TEXT NOT NULL,
  UNIQUE(thread_id, template_id, body),
  FOREIGN KEY(profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS runs (
  run_id TEXT PRIMARY KEY,
  started_at TEXT NOT NULL,
  ended_at TEXT,
  counters_json TEXT NOT NULL,
  outcome TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_profiles_company ON profiles(company);
CREATE INDEX IF NOT EXISTS idx_profiles_title ON profiles(title);
CREATE INDEX IF NOT EXISTS idx_profiles_location ON profiles(location);
CREATE INDEX IF NOT EXISTS idx_connection_requests_sent_at ON connection_requests(sent_at);
CREATE INDEX IF NOT EXISTS idx_messages_profile_id ON messages(profile_id);

-- +subspace Down
-- (No down migrations for PoC)
