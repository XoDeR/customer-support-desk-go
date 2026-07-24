CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('customer','agent','admin')),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled','invited')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX users_email_idx ON users(email);

CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  replaced_by UUID REFERENCES refresh_tokens(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX refresh_tokens_hash_idx ON refresh_tokens(token_hash);

CREATE TABLE teams (
  id UUID PRIMARY KEY, name TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE team_members (
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY(team_id,user_id)
);

CREATE TABLE sla_policies (
  priority TEXT PRIMARY KEY CHECK (priority IN ('low','medium','high','urgent')),
  duration_seconds BIGINT NOT NULL CHECK(duration_seconds > 0),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO sla_policies(priority,duration_seconds) VALUES
 ('low',259200),('medium',86400),('high',28800),('urgent',7200);

CREATE TABLE tickets (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  customer_id UUID NOT NULL REFERENCES users(id),
  assignee_id UUID REFERENCES users(id),
  team_id UUID REFERENCES teams(id),
  status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','pending','resolved','closed')),
  priority TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low','medium','high','urgent')),
  category TEXT NOT NULL DEFAULT 'other' CHECK(category IN ('billing','technical','account','other')),
  sla_due_at TIMESTAMPTZ,
  sla_paused_at TIMESTAMPTZ,
  sla_remaining_seconds BIGINT,
  breached_at TIMESTAMPTZ,
  search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(description,''))) STORED,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX tickets_customer_idx ON tickets(customer_id);
CREATE INDEX tickets_queue_idx ON tickets(status,priority,created_at);
CREATE INDEX tickets_search_idx ON tickets USING GIN(search_vector);

CREATE TABLE comments (
  id UUID PRIMARY KEY, ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  author_id UUID NOT NULL REFERENCES users(id), body TEXT NOT NULL,
  visibility TEXT NOT NULL DEFAULT 'public' CHECK(visibility IN ('public','internal')),
  deleted_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX comments_ticket_idx ON comments(ticket_id,created_at);

CREATE TABLE ticket_status_history (
  id UUID PRIMARY KEY, ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  actor_id UUID NOT NULL REFERENCES users(id), from_status TEXT, to_status TEXT NOT NULL,
  reason TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE audit_events (
  id UUID PRIMARY KEY, ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  actor_id UUID REFERENCES users(id), event_type TEXT NOT NULL, payload JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX audit_ticket_idx ON audit_events(ticket_id,created_at);

CREATE TABLE attachments (
  id UUID PRIMARY KEY, ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  uploaded_by UUID NOT NULL REFERENCES users(id), filename TEXT NOT NULL, storage_key TEXT NOT NULL UNIQUE,
  mime_type TEXT NOT NULL, size_bytes BIGINT NOT NULL CHECK(size_bytes >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX attachments_ticket_idx ON attachments(ticket_id);

CREATE TABLE canned_replies (
  id UUID PRIMARY KEY, title TEXT NOT NULL, body TEXT NOT NULL, team_id UUID REFERENCES teams(id),
  created_by UUID NOT NULL REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE tags (
  id UUID PRIMARY KEY, name TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE ticket_tags (
  ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE, PRIMARY KEY(ticket_id,tag_id)
);
CREATE TABLE saved_filters (
  id UUID PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL, query JSONB NOT NULL DEFAULT '{}', created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
