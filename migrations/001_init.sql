CREATE TABLE IF NOT EXISTS volunteer_state (
    singleton boolean PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    version bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS communities (
    id text PRIMARY KEY,
    name text NOT NULL,
    address text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS service_areas (
    id text PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
    id text PRIMARY KEY,
    display_name text NOT NULL,
    phone text NOT NULL UNIQUE,
    role text NOT NULL CHECK (role IN ('resident','member','captain','admin')),
    active boolean NOT NULL DEFAULT TRUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS teams (
    id text PRIMARY KEY,
    community_id text NOT NULL REFERENCES communities(id),
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    status text NOT NULL CHECK (status IN ('recruiting','paused','closed')),
    visibility text NOT NULL CHECK (visibility IN ('public','community','private')),
    invite_code text NOT NULL UNIQUE,
    captain_id text NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS memberships (
    id text PRIMARY KEY,
    team_id text NOT NULL REFERENCES teams(id),
    user_id text NOT NULL REFERENCES users(id),
    role text NOT NULL CHECK (role IN ('resident','member','captain','admin')),
    status text NOT NULL CHECK (status IN ('pending','active','paused','exited')),
    joined_at timestamptz NOT NULL,
    exited_at timestamptz,
    UNIQUE (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS membership_events (
    id text PRIMARY KEY,
    team_id text NOT NULL REFERENCES teams(id),
    user_id text NOT NULL,
    action text NOT NULL,
    from_value text NOT NULL DEFAULT '',
    to_value text NOT NULL DEFAULT '',
    actor_id text NOT NULL REFERENCES users(id),
    reason text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS activities (
    id text PRIMARY KEY,
    team_id text NOT NULL REFERENCES teams(id),
    title text NOT NULL,
    description text NOT NULL DEFAULT '',
    status text NOT NULL CHECK (status IN ('draft','published','cancelled','completed')),
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL CHECK (end_at > start_at),
    location text NOT NULL,
    created_by text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    cancelled_at timestamptz
);

CREATE TABLE IF NOT EXISTS role_slots (
    id text PRIMARY KEY,
    activity_id text NOT NULL REFERENCES activities(id),
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    capacity integer NOT NULL CHECK (capacity > 0),
    status text NOT NULL CHECK (status IN ('open','filled','cancelled'))
);

CREATE TABLE IF NOT EXISTS claims (
    id text PRIMARY KEY,
    slot_id text NOT NULL REFERENCES role_slots(id),
    activity_id text NOT NULL REFERENCES activities(id),
    user_id text NOT NULL,
    status text NOT NULL CHECK (status IN ('active','waitlisted','cancelled','checked_in','no_show')),
    claimed_at timestamptz NOT NULL,
    checked_at timestamptz
);

CREATE TABLE IF NOT EXISTS service_records (
    id text PRIMARY KEY,
    supersedes_id text REFERENCES service_records(id),
    claim_id text NOT NULL REFERENCES claims(id),
    activity_id text NOT NULL REFERENCES activities(id),
    user_id text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    status text NOT NULL CHECK (status IN ('started','ended','pending_review','confirmed','correction_open')),
    started_at timestamptz,
    ended_at timestamptz,
    duration_minutes integer NOT NULL DEFAULT 0 CHECK (duration_minutes >= 0 AND duration_minutes <= 1440),
    photo_note text NOT NULL DEFAULT '',
    reviewer_id text NOT NULL DEFAULT '',
    correction_note text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (claim_id, version)
);

CREATE TABLE IF NOT EXISTS risks (
    id text PRIMARY KEY,
    activity_id text NOT NULL REFERENCES activities(id),
    team_id text NOT NULL REFERENCES teams(id),
    title text NOT NULL,
    severity text NOT NULL CHECK (severity IN ('low','medium','high')),
    description text NOT NULL,
    owner_id text NOT NULL,
    status text NOT NULL CHECK (status IN ('open','in_progress','resolved')),
    due_at timestamptz NOT NULL,
    resolved_at timestamptz
);

CREATE TABLE IF NOT EXISTS follow_ups (
    id text PRIMARY KEY,
    risk_id text NOT NULL REFERENCES risks(id),
    title text NOT NULL,
    owner_id text NOT NULL,
    due_at timestamptz NOT NULL,
    done boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS handoffs (
    id text PRIMARY KEY,
    activity_id text NOT NULL REFERENCES activities(id),
    risk_id text NOT NULL REFERENCES risks(id),
    from_user_id text NOT NULL,
    to_user_id text NOT NULL,
    status text NOT NULL CHECK (status IN ('pending','acknowledged','overdue','completed')),
    note text NOT NULL,
    due_at timestamptz NOT NULL,
    ack_at timestamptz
);

CREATE TABLE IF NOT EXISTS announcements (
    id text PRIMARY KEY,
    team_id text NOT NULL REFERENCES teams(id),
    author_id text NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notices (
    id text PRIMARY KEY,
    user_id text NOT NULL,
    kind text NOT NULL,
    message text NOT NULL,
    read boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_events (
    id text PRIMARY KEY,
    actor_id text NOT NULL,
    entity_type text NOT NULL DEFAULT '',
    entity_id text NOT NULL DEFAULT '',
    action text NOT NULL,
    details text NOT NULL DEFAULT '',
    request_id text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS attachments (
    id text PRIMARY KEY,
    owner_id text NOT NULL,
    relative_path text NOT NULL,
    media_type text NOT NULL,
    byte_size bigint NOT NULL CHECK (byte_size > 0),
    sha256 text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_activities_team_window ON activities(team_id, start_at, end_at);
CREATE INDEX IF NOT EXISTS idx_claims_slot_status ON claims(slot_id, status);
CREATE INDEX IF NOT EXISTS idx_claims_user_activity ON claims(user_id, activity_id, status);
CREATE INDEX IF NOT EXISTS idx_handoffs_due_status ON handoffs(status, due_at);
CREATE INDEX IF NOT EXISTS idx_service_records_user_status ON service_records(user_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_entity_time ON audit_events(entity_type, entity_id, created_at DESC);
