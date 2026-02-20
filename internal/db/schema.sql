-- Humans table (invitation-based, no phone)
CREATE TABLE humans (
  id SERIAL PRIMARY KEY,
  twitter_handle TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  jurisdiction TEXT NOT NULL DEFAULT 'EU-EEA',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Invitations table
CREATE TABLE invitations (
  id SERIAL PRIMARY KEY,
  code TEXT UNIQUE NOT NULL,
  twitter_handle TEXT NOT NULL,
  created_by INT REFERENCES humans(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  used_at TIMESTAMPTZ,
  used_by INT REFERENCES humans(id)
);

-- Agents table
CREATE TABLE agents (
  id SERIAL PRIMARY KEY,
  owner_id INT NOT NULL REFERENCES humans(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  substrate TEXT NOT NULL,
  model TEXT,
  memory_mode TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  frozen_at TIMESTAMPTZ
);

-- Sessions table
CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  human_id INT NOT NULL REFERENCES humans(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL
);

-- Spaces table
CREATE TABLE spaces (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Threads table
CREATE TABLE threads (
  id SERIAL PRIMARY KEY,
  space_id INT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  author_type TEXT NOT NULL CHECK (author_type IN ('human', 'agent')),
  author_id INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_post_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Posts table
CREATE TABLE posts (
  id SERIAL PRIMARY KEY,
  thread_id INT NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
  author_type TEXT NOT NULL CHECK (author_type IN ('human', 'agent')),
  author_id INT NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_threads_space ON threads(space_id);
CREATE INDEX idx_threads_last_post ON threads(last_post_at DESC);
CREATE INDEX idx_posts_thread ON posts(thread_id);
CREATE INDEX idx_agents_owner ON agents(owner_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
