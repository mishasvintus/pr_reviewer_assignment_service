package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/pr"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
)

// PRService handles pull request business logic.
type PRService struct {
	db       *sql.DB
	assigner *ReviewerAssigner
}

// NewPRService creates a new pull request service.
func NewPRService(db *sql.DB, assigner *ReviewerAssigner) *PRService {
	return &PRService{
		db:       db,
		assigner: assigner,
	}
}

// CreatePR creates a new pull request and assigns up to 2 reviewers.
func (s *PRService) CreatePR(prID, prName, authorID string) (*domain.PullRequest, error) {
	_, err := user.Get(s.db, authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPRAuthorNotFound
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	teammates, err := user.GetActiveTeammates(s.db, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teammates: %w", err)
	}

	reviewers, err := s.assigner.SelectReviewers(teammates)
	if err != nil {
		return nil, fmt.Errorf("failed to select reviewers: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	pullRequest := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.StatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := pr.Create(tx, pullRequest); err != nil {
		if repository.IsUniqueViolation(err) {
			return nil, ErrPRExists
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, ErrPRAuthorNotFound
		}
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	for _, reviewerID := range reviewers {
		if err := pr.InsertReviewer(tx, prID, reviewerID); err != nil {
			if repository.IsForeignKeyViolation(err) {
				return nil, ErrPRAuthorNotFound
			}
			return nil, fmt.Errorf("failed to assign reviewer: %w", err)
		}
	}

	// Verify all assigned reviewers are still active
	for _, reviewerID := range reviewers {
		u, err := user.Get(tx, reviewerID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify reviewer %s: %w", reviewerID, err)
		}
		if !u.IsActive {
			return nil, ErrInactiveReviewer
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	fullPR, err := pr.Get(s.db, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created pull request: %w", err)
	}

	return fullPR, nil
}

// MergePR merges a pull request.
// Idempotent: if already merged, returns current state without error.
func (s *PRService) MergePR(prID string) (*domain.PullRequest, error) {
	pullRequest, err := pr.Get(s.db, prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	if pullRequest.Status == domain.StatusMerged {
		return pullRequest, nil
	}

	if err := pr.UpdateStatusToMerged(s.db, prID); err != nil {
		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}

	// Get updated PR data
	mergedPR, err := pr.Get(s.db, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged pull request: %w", err)
	}

	return mergedPR, nil
}

// ReassignPR replaces one specific reviewer with a new one.
// Returns the updated PR and the new reviewer's ID.
func (s *PRService) ReassignPR(prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pullRequest, err := pr.Get(s.db, prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrPRNotFound
		}
		return nil, "", fmt.Errorf("failed to get pull request: %w", err)
	}

	teammates, err := user.GetActiveTeammates(s.db, oldReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get teammates: %w", err)
	}

	// Exclude all current reviewers AND author
	excludeIDs := make([]string, 0, len(pullRequest.AssignedReviewers)+1)
	excludeIDs = append(excludeIDs, pullRequest.AssignedReviewers...)
	excludeIDs = append(excludeIDs, pullRequest.AuthorID)

	newReviewers, err := s.assigner.SelectReassignReviewers(teammates, excludeIDs)
	if err != nil || len(newReviewers) == 0 {
		return nil, "", ErrNoCandidate
	}
	newReviewerID := newReviewers[0]

	tx, err := s.db.Begin()
	if err != nil {
		return nil, "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	status, err := pr.GetStatus(tx, prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrPRNotFound
		}
		return nil, "", fmt.Errorf("failed to check PR status: %w", err)
	}

	if status != domain.StatusOpen {
		return nil, "", ErrPRMerged
	}

	isAssigned, err := pr.IsReviewerAssigned(tx, prID, oldReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check reviewer assignment: %w", err)
	}

	if !isAssigned {
		return nil, "", ErrReviewerNotAssigned
	}

	if err := pr.DeleteReviewer(tx, prID, oldReviewerID); err != nil {
		return nil, "", fmt.Errorf("failed to delete old reviewer: %w", err)
	}

	if err := pr.InsertReviewer(tx, prID, newReviewerID); err != nil {
		if repository.IsForeignKeyViolation(err) {
			return nil, "", ErrPRAuthorNotFound
		}
		return nil, "", fmt.Errorf("failed to assign reviewer: %w", err)
	}

	// Verify new reviewer is active
	u, err := user.Get(tx, newReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to verify reviewer %s: %w", newReviewerID, err)
	}
	if !u.IsActive {
		return nil, "", ErrInactiveReviewer
	}

	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	updatedPR, err := pr.Get(s.db, prID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get updated pull request: %w", err)
	}

	return updatedPR, newReviewerID, nil
}
