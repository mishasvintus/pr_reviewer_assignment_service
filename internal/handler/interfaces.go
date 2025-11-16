package handler

import (
	"github.com/mishasvintus/avito_backend_internship/internal/domain"
)

// TeamServiceInterface defines the interface for team operations.
type TeamServiceInterface interface {
	CreateTeam(teamName string, members []domain.TeamMember) error
	GetTeam(teamName string) (*domain.Team, error)
	DeactivateTeam(teamName string) error
}

// UserServiceInterface defines the interface for user operations.
type UserServiceInterface interface {
	SetIsActive(userID string, isActive bool) (*domain.User, error)
	GetUserReviews(userID string) ([]domain.PullRequestShort, error)
}

// PRServiceInterface defines the interface for pull request operations.
type PRServiceInterface interface {
	CreatePR(prID, prName, authorID string) (*domain.PullRequest, error)
	MergePR(prID string) (*domain.PullRequest, error)
	ReassignPR(prID, oldReviewerID string) (*domain.PullRequest, string, error)
}
