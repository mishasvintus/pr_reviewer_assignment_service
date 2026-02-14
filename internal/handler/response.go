package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode represents error codes from OpenAPI spec.
type ErrorCode string

const (
	ErrorTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorPRExists    ErrorCode = "PR_EXISTS"
	ErrorPRMerged    ErrorCode = "PR_MERGED"
	ErrorNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorNotFound    ErrorCode = "NOT_FOUND"
)

// ErrorResponse represents error response structure.
type ErrorResponse struct {
	Error struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	} `json:"error"`
}

// SuccessResponse represents success response structure.
type SuccessResponse struct {
	Team *TeamResponse `json:"team,omitempty"`
	User *UserResponse `json:"user,omitempty"`
	PR   *PRResponse   `json:"pr,omitempty"`
}

// TeamResponse wraps team data.
type TeamResponse struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

// TeamMember represents a team member in response.
type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// UserResponse wraps user data.
type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// PRResponse wraps pull request data.
type PRResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	TeamName          string   `json:"team_name"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         string   `json:"createdAt,omitempty"`
	MergedAt          string   `json:"mergedAt,omitempty"`
}

// ReassignResponse wraps reassign response.
type ReassignResponse struct {
	PR         *PRResponse `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}

// GetReviewResponse wraps get review response.
type GetReviewResponse struct {
	UserID       string            `json:"user_id"`
	PullRequests []PRShortResponse `json:"pull_requests"`
}

// PRShortResponse represents short PR in response.
type PRShortResponse struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	TeamName        string `json:"team_name"`
	Status          string `json:"status"`
}

// StatisticsResponse wraps statistics response.
type StatisticsResponse struct {
	Overall struct {
		TotalPRs         int64 `json:"total_prs"`
		TotalAssignments int64 `json:"total_assignments"`
		TotalUsers       int64 `json:"total_users"`
		TotalTeams       int64 `json:"total_teams"`
	} `json:"overall"`
	ReviewerStats []ReviewerStatResponse `json:"reviewer_stats"`
	AuthorStats   []AuthorStatResponse   `json:"author_stats"`
}

// ReviewerStatResponse represents reviewer statistics in response.
type ReviewerStatResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

// AuthorStatResponse represents author statistics in response.
type AuthorStatResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

// Error sends error response.
func Error(c *gin.Context, code ErrorCode, message string, statusCode int) {
	c.JSON(statusCode, ErrorResponse{
		Error: struct {
			Code    ErrorCode `json:"code"`
			Message string    `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

// NotFound sends 404 error.
func NotFound(c *gin.Context, message string) {
	Error(c, ErrorNotFound, message, http.StatusNotFound)
}

// Conflict sends 409 error.
func Conflict(c *gin.Context, code ErrorCode, message string) {
	Error(c, code, message, http.StatusConflict)
}

// BadRequest sends 400 error.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: struct {
			Code    ErrorCode `json:"code"`
			Message string    `json:"message"`
		}{
			Code:    "",
			Message: message,
		},
	})
}

// InternalError sends 500 error.
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: struct {
			Code    ErrorCode `json:"code"`
			Message string    `json:"message"`
		}{
			Code:    "",
			Message: message,
		},
	})
}
