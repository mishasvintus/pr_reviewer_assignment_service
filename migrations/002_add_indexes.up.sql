-- team.Get() - WHERE team_name = $1
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);

-- user.GetActiveTeammates() - JOIN and filter by is_active
CREATE INDEX IF NOT EXISTS idx_users_team_name_is_active ON users(team_name, is_active);

-- pr.Get() - WHERE pull_request_id = $1
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pull_request_id ON pr_reviewers(pull_request_id);

-- pr.GetByUser() - JOIN and WHERE rev.user_id = $1
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);

-- pr.GetByUser() - ORDER BY pr.created_at DESC
CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests(created_at DESC);

-- Reassign by PR's team (GetActiveByTeam / lookup by team_name)
CREATE INDEX IF NOT EXISTS idx_pull_requests_team_name ON pull_requests(team_name);

