package pr

import (
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// GetOpenPRsWithReviewersFromTeam returns open PRs that have at least one reviewer from the specified team.
// Map: prID -> list of reviewer IDs from that team.
func GetOpenPRsWithReviewersFromTeam(exec repository.DBTX, teamName string) (map[string][]string, error) {
	query := `
		SELECT pr.pull_request_id, rev.user_id
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

	byPR := make(map[string][]string)
	for rows.Next() {
		var prID, reviewerID string
		if err := rows.Scan(&prID, &reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		byPR[prID] = append(byPR[prID], reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return byPR, nil
}
