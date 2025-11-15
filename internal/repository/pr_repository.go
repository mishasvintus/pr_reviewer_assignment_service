package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/mishasvintus/avito_backend_internship/internal/domain"
)

// PRRepository handles pull request database operations.
type PRRepository struct {
	db *sql.DB
}

// NewPRRepository creates a new pull request repository.
func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}

// Create creates a new pull request with assigned reviewers in a transaction.
func (r *PRRepository) Create(pr *domain.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert pull request
	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	_, err = tx.Exec(query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, now)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	// Insert reviewers
	reviewerQuery := `
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
	`
	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.Exec(reviewerQuery, pr.PullRequestID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	// Verify all assigned reviewers are still active
	if len(pr.AssignedReviewers) > 0 {
		checkQuery := `
			SELECT user_id 
			FROM users 
			WHERE user_id = ANY($1) AND is_active = false
		`
		rows, err := tx.Query(checkQuery, pq.Array(pr.AssignedReviewers))
		if err != nil {
			return fmt.Errorf("failed to verify reviewers: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			var inactiveUserID string
			rows.Scan(&inactiveUserID)
			return &ErrInactiveReviewer{UserID: inactiveUserID}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Get retrieves a pull request by ID with all assigned reviewers.
func (r *PRRepository) Get(prID string) (*domain.PullRequest, error) {
	// Get PR details
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`
	var pr domain.PullRequest
	err := r.db.QueryRow(query).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
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
	rows, err := r.db.Query(reviewersQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

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

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

// GetByUser retrieves all pull requests assigned to a user for review.
func (r *PRRepository) GetByUser(userID string) ([]domain.PullRequestShort, error) {
	query := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pull requests: %w", err)
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return prs, nil
}

// Merge updates the pull request status to MERGED and sets merged_at timestamp.
func (r *PRRepository) Merge(prID string) error {
	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2
		WHERE pull_request_id = $3 AND status = $4
	`
	now := time.Now()
	result, err := r.db.Exec(query, domain.StatusMerged, now, prID, domain.StatusOpen)
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
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

// Reassign replaces assigned reviewers for a pull request.
// Verifies that PR is not merged and requester is currently assigned.
func (r *PRRepository) Reassign(prID, requesterID string, newReviewers []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check PR status (must be OPEN)
	var status string
	statusQuery := `SELECT status FROM pull_requests WHERE pull_request_id = $1`
	err = tx.QueryRow(statusQuery, prID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("failed to check PR status: %w", err)
	}

	if status != "OPEN" {
		return fmt.Errorf("cannot reassign merged pull request")
	}

	// Check requester is currently assigned
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)`
	err = tx.QueryRow(checkQuery, prID, requesterID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check reviewer assignment: %w", err)
	}

	if !exists {
		return fmt.Errorf("user %s is not assigned to this pull request", requesterID)
	}

	// Delete old reviewers
	deleteQuery := `DELETE FROM pr_reviewers WHERE pull_request_id = $1`
	_, err = tx.Exec(deleteQuery, prID)
	if err != nil {
		return fmt.Errorf("failed to delete old reviewers: %w", err)
	}

	// Insert new reviewers
	insertQuery := `INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`
	for _, reviewerID := range newReviewers {
		_, err := tx.Exec(insertQuery, prID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to assign new reviewer %s: %w", reviewerID, err)
		}
	}

	// Verify new reviewers are active
	if len(newReviewers) > 0 {
		checkActiveQuery := `
			SELECT user_id 
			FROM users 
			WHERE user_id = ANY($1) AND is_active = false
		`
		rows, err := tx.Query(checkActiveQuery, pq.Array(newReviewers))
		if err != nil {
			return fmt.Errorf("failed to verify reviewers: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			var inactiveUserID string
			rows.Scan(&inactiveUserID)
			return &ErrInactiveReviewer{UserID: inactiveUserID}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Exists checks if a pull request exists.
func (r *PRRepository) Exists(prID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
	err := r.db.QueryRow(query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pull request existence: %w", err)
	}
	return exists, nil
}

// GetStatus returns the status of a pull request.
func (r *PRRepository) GetStatus(prID string) (domain.PRStatus, error) {
	var status domain.PRStatus
	query := `SELECT status FROM pull_requests WHERE pull_request_id = $1`
	err := r.db.QueryRow(query, prID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", fmt.Errorf("failed to get pull request status: %w", err)
	}
	return status, nil
}
