package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// UserService handles user business logic.
type UserService struct {
	userRepo *repository.UserRepository
	prRepo   *repository.PRRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo *repository.UserRepository, prRepo *repository.PRRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

// SetIsActive updates the is_active status of a user.
func (s *UserService) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.SetIsActive(userID, isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return user, nil
}

// GetUserReviews returns all pull requests where the user is assigned as a reviewer.
// Only returns OPEN pull requests.
func (s *UserService) GetUserReviews(userID string) ([]domain.PullRequestShort, error) {
	prs, err := s.prRepo.GetByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reviews: %w", err)
	}

	openPRs := make([]domain.PullRequestShort, 0)
	for _, pr := range prs {
		if pr.Status == domain.StatusOpen {
			openPRs = append(openPRs, pr)
		}
	}

	return openPRs, nil
}
