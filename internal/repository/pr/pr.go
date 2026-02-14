package pr

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// ErrReviewerNotAssigned is returned when the reviewer to replace is not assigned to the PR.
var ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")

// Create inserts a new pull request.
func Create(exec repository.DBTX, pr *domain.PullRequest) error {
	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, team_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now()
	_, err := exec.Exec(query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.TeamName, pr.Status, now)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	return nil
}

// InsertReviewer assigns a reviewer to a pull request.
func InsertReviewer(exec repository.DBTX, prID, userID string) error {
	query := `INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`
	_, err := exec.Exec(query, prID, userID)
	if err != nil {
		return fmt.Errorf("failed to insert reviewer: %w", err)
	}
	return nil
}

// Get retrieves a pull request by ID with all assigned reviewers.
func Get(exec repository.DBTX, prID string) (*domain.PullRequest, error) {
	// Get PR details
	query := `
		SELECT pull_request_id, pull_request_name, author_id, team_name, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`
	var p domain.PullRequest
	err := exec.QueryRow(query, prID).Scan(
		&p.PullRequestID,
		&p.PullRequestName,
		&p.AuthorID,
		&p.TeamName,
		&p.Status,
		&p.CreatedAt,
		&p.MergedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	// Get assigned reviewers
	reviewersQuery := `
		SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
	`
	rows, err := exec.Query(reviewersQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	p.AssignedReviewersIDs = reviewers
	return &p, nil
}

// GetByUser retrieves all pull requests assigned to a user for review.
func GetByUser(exec repository.DBTX, userID string) ([]domain.PullRequestShort, error) {
	query := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.team_name, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC
	`
	rows, err := exec.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pull requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var p domain.PullRequestShort
		if err := rows.Scan(&p.PullRequestID, &p.PullRequestName, &p.AuthorID, &p.TeamName, &p.Status); err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}
		prs = append(prs, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return prs, nil
}

// UpdateStatusToMerged updates the pull request status to MERGED.
// Returns sql.ErrNoRows if PR doesn't exist or already merged.
func UpdateStatusToMerged(exec repository.DBTX, prID string) error {
	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2
		WHERE pull_request_id = $3 AND status = $4
	`
	now := time.Now()
	result, err := exec.Exec(query, domain.StatusMerged, now, prID, domain.StatusOpen)
	if err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteReviewer removes a specific reviewer from a pull request.
func DeleteReviewer(exec repository.DBTX, prID, userID string) error {
	query := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`
	_, err := exec.Exec(query, prID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete reviewer: %w", err)
	}
	return nil
}

// ReplaceReviewer atomically replaces oldReviewerID with newReviewerID for the given PR.
// Returns ErrReviewerNotAssigned if oldReviewerID was not assigned to this PR.
func ReplaceReviewer(exec repository.DBTX, prID, oldReviewerID, newReviewerID string) error {
	query := `
		WITH deleted AS (
			DELETE FROM pr_reviewers
			WHERE pull_request_id = $1 AND user_id = $2
			RETURNING pull_request_id
		)
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		SELECT $1, $3 FROM deleted
	`
	result, err := exec.Exec(query, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to replace reviewer: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return ErrReviewerNotAssigned
	}
	return nil
}

// GetStatus returns the status of a pull request.
func GetStatus(exec repository.DBTX, prID string) (domain.PRStatus, error) {
	var status domain.PRStatus
	query := `SELECT status FROM pull_requests WHERE pull_request_id = $1`
	err := exec.QueryRow(query, prID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", fmt.Errorf("failed to get pull request status: %w", err)
	}
	return status, nil
}

