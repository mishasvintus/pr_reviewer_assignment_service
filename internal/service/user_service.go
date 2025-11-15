package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/pr"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
)

// UserService handles user business logic.
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new user service.
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// SetIsActive updates the is_active status of a user.
func (s *UserService) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	u, err := user.SetIsActive(s.db, userID, isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return u, nil
}

// GetUserReviews returns all pull requests where the user is assigned as a reviewer.
func (s *UserService) GetUserReviews(userID string) ([]domain.PullRequestShort, error) {
	prs, err := pr.GetByUser(s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reviews: %w", err)
	}

	return prs, nil
}
