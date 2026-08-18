CREATE UNIQUE INDEX IF NOT EXISTS uq_active_captain_per_team
    ON memberships(team_id) WHERE role = 'captain' AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_membership_events_team_user ON membership_events(team_id, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risks_team_due ON risks(team_id, status, due_at);
