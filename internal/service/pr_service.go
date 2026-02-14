package service

import (
	"database/sql"
	"errors"
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
	author, err := user.Get(s.db, authorID)
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
	defer func() { _ = tx.Rollback() }()

	pullRequest := &domain.PullRequest{
		PullRequestID:        prID,
		PullRequestName:      prName,
		AuthorID:             authorID,
		TeamName:             author.TeamName,
		Status:               domain.StatusOpen,
		AssignedReviewersIDs: reviewers,
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

const maxReviewers = 2

// ReplenishReviewers ensures the PR has up to maxReviewers reviewers from its team.
// Does nothing if PR already has >= maxReviewers or is not OPEN.
func (s *PRService) ReplenishReviewers(exec repository.DBTX, prID string) error {
	pullRequest, err := pr.Get(exec, prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to get PR: %w", err)
	}
	if pullRequest.Status != domain.StatusOpen {
		return nil
	}
	reviewerCount := len(pullRequest.AssignedReviewersIDs)
	if reviewerCount >= maxReviewers {
		return nil
	}

	candidates, err := user.GetActiveByTeam(exec, pullRequest.TeamName)
	if err != nil {
		return fmt.Errorf("failed to get active users in PR team: %w", err)
	}
	newReviewers, err := s.assigner.SelectReassignReviewers(candidates, pullRequest.AuthorID, pullRequest.AssignedReviewersIDs)
	if err != nil || len(newReviewers) == 0 {
		return nil
	}

	need := min(maxReviewers - reviewerCount, len(newReviewers))

	newReviewers = newReviewers[:need]

	for _, reviewer := range newReviewers {
		if err := pr.InsertReviewer(exec, prID, reviewer); err != nil {
			return fmt.Errorf("failed to insert reviewer: %w", err)
		}
	}
	return nil
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
// New reviewer is chosen from the PR's responsible team (team_name).
// Returns the updated PR and the new reviewer's ID.
func (s *PRService) ReassignPR(prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pullRequest, err := pr.Get(s.db, prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrPRNotFound
		}
		return nil, "", fmt.Errorf("failed to get pull request: %w", err)
	}

	candidates, err := user.GetActiveByTeam(s.db, pullRequest.TeamName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get active users in PR team: %w", err)
	}

	newReviewers, err := s.assigner.SelectReassignReviewers(candidates, pullRequest.AuthorID, pullRequest.AssignedReviewersIDs)
	if err != nil || len(newReviewers) == 0 {
		return nil, "", ErrNoCandidate
	}
	newReviewerID := newReviewers[0]

	tx, err := s.db.Begin()
	if err != nil {
		return nil, "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

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

	if err := pr.ReplaceReviewer(tx, prID, oldReviewerID, newReviewerID); err != nil {
		if errors.Is(err, pr.ErrReviewerNotAssigned) {
			return nil, "", ErrReviewerNotAssigned
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, "", ErrPRAuthorNotFound
		}
		return nil, "", fmt.Errorf("failed to replace reviewer: %w", err)
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
