package pr

import (
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// PRWithReviewer represents a PR with one of its reviewers.
type PRWithReviewer struct {
	PullRequestID string
	AuthorID      string
	ReviewerID    string
}

// GetOpenPRsWithReviewersFromTeam returns all open PRs that have at least one reviewer from the specified team.
func GetOpenPRsWithReviewersFromTeam(exec repository.DBTX, teamName string) ([]PRWithReviewer, error) {
	query := `
		SELECT DISTINCT pr.pull_request_id, pr.author_id, rev.user_id as reviewer_id
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		JOIN users u ON rev.user_id = u.user_id
		WHERE pr.status = 'OPEN' AND u.team_name = $1
	`
	rows, err := exec.Query(query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get open PRs with reviewers from team: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []PRWithReviewer
	for rows.Next() {
		var item PRWithReviewer
		if err := rows.Scan(&item.PullRequestID, &item.AuthorID, &item.ReviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan PR with reviewer: %w", err)
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}
