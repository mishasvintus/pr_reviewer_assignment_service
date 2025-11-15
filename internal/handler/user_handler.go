package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mishasvintus/avito_backend_internship/internal/service"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// SetIsActive handles POST /users/setIsActive.
func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		IsActive bool   `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	user, err := h.userService.SetIsActive(req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			NotFound(c, "user not found")
			return
		}
		InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		User: &UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	})
}

// GetReview handles GET /users/getReview.
func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		BadRequest(c, "user_id parameter is required")
		return
	}

	prs, err := h.userService.GetUserReviews(userID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	// Convert domain.PullRequestShort to response format
	prResponses := make([]PRShortResponse, len(prs))
	for i, p := range prs {
		prResponses[i] = PRShortResponse{
			PullRequestID:   p.PullRequestID,
			PullRequestName: p.PullRequestName,
			AuthorID:        p.AuthorID,
			Status:          string(p.Status),
		}
	}

	c.JSON(http.StatusOK, GetReviewResponse{
		UserID:       userID,
		PullRequests: prResponses,
	})
}
