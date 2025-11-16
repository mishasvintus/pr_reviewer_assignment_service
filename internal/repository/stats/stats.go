package stats

import (
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// ReviewerStat represents statistics for a reviewer.
type ReviewerStat struct {
	UserID   string
	Username string
	Count    int64
}

// AuthorStat represents statistics for an author.
type AuthorStat struct {
	UserID   string
	Username string
	Count    int64
}

// OverallStats represents overall statistics.
type OverallStats struct {
	TotalPRs         int64
	TotalAssignments int64
	TotalUsers       int64
	TotalTeams       int64
}

// GetReviewerStats returns statistics about reviewer assignments per user.
func GetReviewerStats(exec repository.DBTX) ([]ReviewerStat, error) {
	query := `
		SELECT u.user_id, u.username, COUNT(pr.user_id) as assignment_count
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
		GROUP BY u.user_id, u.username
		ORDER BY assignment_count DESC, u.user_id
	`
	rows, err := exec.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []ReviewerStat
	for rows.Next() {
		var stat ReviewerStat
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}

// GetAuthorStats returns statistics about PRs created per author.
func GetAuthorStats(exec repository.DBTX) ([]AuthorStat, error) {
	query := `
		SELECT u.user_id, u.username, COUNT(pr.pull_request_id) as pr_count
		FROM users u
		LEFT JOIN pull_requests pr ON u.user_id = pr.author_id
		GROUP BY u.user_id, u.username
		ORDER BY pr_count DESC, u.user_id
	`
	rows, err := exec.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get author stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []AuthorStat
	for rows.Next() {
		var stat AuthorStat
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.Count); err != nil {
			return nil, fmt.Errorf("failed to scan author stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}

// GetOverallStats returns overall statistics.
func GetOverallStats(exec repository.DBTX) (*OverallStats, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM pull_requests) as total_prs,
			(SELECT COUNT(*) FROM pr_reviewers) as total_assignments,
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM teams) as total_teams
	`
	var stats OverallStats
	err := exec.QueryRow(query).Scan(
		&stats.TotalPRs,
		&stats.TotalAssignments,
		&stats.TotalUsers,
		&stats.TotalTeams,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	return &stats, nil
}
