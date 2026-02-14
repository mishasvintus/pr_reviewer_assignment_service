-- Drop indexes

DROP INDEX IF EXISTS idx_pull_requests_team_name;
DROP INDEX IF EXISTS idx_pull_requests_created_at;
DROP INDEX IF EXISTS idx_pr_reviewers_user_id;
DROP INDEX IF EXISTS idx_pr_reviewers_pull_request_id;
DROP INDEX IF EXISTS idx_users_team_name_is_active;
DROP INDEX IF EXISTS idx_users_team_name;

