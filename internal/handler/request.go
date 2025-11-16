package handler

import "github.com/mishasvintus/avito_backend_internship/internal/domain"

// CreatePRRequest represents request body for POST /pullRequest/create.
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

// MergePRRequest represents request body for POST /pullRequest/merge.
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

// ReassignPRRequest represents request body for POST /pullRequest/reassign.
type ReassignPRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_user_id" binding:"required"`
}

// AddTeamRequest represents request body for POST /team/add.
type AddTeamRequest struct {
	TeamName string              `json:"team_name" binding:"required"`
	Members  []domain.TeamMember `json:"members" binding:"required"`
}

// DeactivateTeamRequest represents request body for POST /team/deactivate.
type DeactivateTeamRequest struct {
	TeamName string `json:"team_name" binding:"required"`
}

// SetIsActiveRequest represents request body for POST /users/setIsActive.
type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive *bool  `json:"is_active" binding:"required"`
}
