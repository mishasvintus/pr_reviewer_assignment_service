package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// PRService handles pull request business logic.
type PRService struct {
	prRepo   *repository.PRRepository
	userRepo *repository.UserRepository
	assigner *ReviewerAssigner
}

// NewPRService creates a new pull request service.
func NewPRService(prRepo *repository.PRRepository, userRepo *repository.UserRepository, assigner *ReviewerAssigner) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		assigner: assigner,
	}
}

// CreatePR creates a new pull request and assigns up to 2 reviewers.
func (s *PRService) CreatePR(prID, prName, authorID string) (*domain.PullRequest, error) {
	_, err := s.userRepo.Get(authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("author not found")
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	teammates, err := s.userRepo.GetActiveTeammates(authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teammates: %w", err)
	}

	reviewers, err := s.assigner.SelectReviewers(teammates)
	if err != nil {
		return nil, fmt.Errorf("failed to select reviewers: %w", err)
	}

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.StatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.prRepo.Create(pr); err != nil {
		if repository.IsUniqueViolation(err) {
			return nil, fmt.Errorf("pull request already exists")
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, fmt.Errorf("author or reviewer not found")
		}
		if repository.IsInactiveReviewer(err) {
			return nil, fmt.Errorf("cannot assign inactive reviewer")
		}
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	fullPR, err := s.prRepo.Get(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created pull request: %w", err)
	}

	return fullPR, nil
}

// MergePR merges a pull request.
// Idempotent: if already merged, returns current state without error.
func (s *PRService) MergePR(prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.Get(prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull request not found")
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	if pr.Status == domain.StatusMerged {
		return pr, nil
	}

	if err := s.prRepo.Merge(prID); err != nil {
		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}

	mergedPR, err := s.prRepo.Get(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged pull request: %w", err)
	}

	return mergedPR, nil
}

// ReassignPR reassigns reviewers for a pull request.
func (s *PRService) ReassignPR(prID, userID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.Get(prID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull request not found")
		}
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	// Get active teammates
	teammates, err := s.userRepo.GetActiveTeammates(pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teammates: %w", err)
	}

	// Select new reviewers
	newReviewers, err := s.assigner.SelectReassignReviewers(teammates, pr.AssignedReviewers)
	if err != nil {
		return nil, fmt.Errorf("no candidates available for reassignment")
	}

	// Reassign with validation in transaction
	err = s.prRepo.Reassign(prID, userID, newReviewers)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pull request not found")
		}
		if repository.IsForeignKeyViolation(err) {
			return nil, fmt.Errorf("reviewer not found")
		}
		if repository.IsInactiveReviewer(err) {
			return nil, fmt.Errorf("cannot assign inactive reviewer")
		}
		return nil, fmt.Errorf("failed to reassign reviewers: %w", err)
	}

	// Get updated PR
	updatedPR, err := s.prRepo.Get(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated pull request: %w", err)
	}

	return updatedPR, nil
}
