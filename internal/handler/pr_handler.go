package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/service"
)

// PRHandler handles pull request-related HTTP requests.
type PRHandler struct {
	prService PRServiceInterface
}

// NewPRHandler creates a new PR handler.
func NewPRHandler(prService PRServiceInterface) *PRHandler {
	return &PRHandler{prService: prService}
}

// CreatePR handles POST /pullRequest/create.
func (h *PRHandler) CreatePR(c *gin.Context) {
	var req CreatePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	pr, err := h.prService.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if errors.Is(err, service.ErrPRExists) {
			Conflict(c, ErrorPRExists, "PR id already exists")
			return
		}
		if errors.Is(err, service.ErrPRAuthorNotFound) {
			NotFound(c, "author or team not found")
			return
		}
		if errors.Is(err, service.ErrInactiveReviewer) {
			BadRequest(c, err.Error())
			return
		}
		InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		PR: domainToPRResponse(pr),
	})
}

// MergePR handles POST /pullRequest/merge.
func (h *PRHandler) MergePR(c *gin.Context) {
	var req MergePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	pr, err := h.prService.MergePR(req.PullRequestID)
	if err != nil {
		if errors.Is(err, service.ErrPRNotFound) {
			NotFound(c, "pull request not found")
			return
		}
		InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		PR: domainToPRResponse(pr),
	})
}

// ReassignPR handles POST /pullRequest/reassign.
func (h *PRHandler) ReassignPR(c *gin.Context) {
	var req ReassignPRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	pr, replacedBy, err := h.prService.ReassignPR(req.PullRequestID, req.OldUserID)
	if err != nil {
		if errors.Is(err, service.ErrPRNotFound) || errors.Is(err, service.ErrPRAuthorNotFound) {
			NotFound(c, "pull request or user not found")
			return
		}
		if errors.Is(err, service.ErrPRMerged) {
			Conflict(c, ErrorPRMerged, "cannot reassign on merged PR")
			return
		}
		if errors.Is(err, service.ErrReviewerNotAssigned) {
			Conflict(c, ErrorNotAssigned, "reviewer is not assigned to this PR")
			return
		}
		if errors.Is(err, service.ErrNoCandidate) {
			Conflict(c, ErrorNoCandidate, "no active replacement candidate in team")
			return
		}
		if errors.Is(err, service.ErrInactiveReviewer) {
			BadRequest(c, err.Error())
			return
		}
		InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, ReassignResponse{
		PR:         domainToPRResponse(pr),
		ReplacedBy: replacedBy,
	})
}

// domainToPRResponse converts domain.PullRequest to PRResponse.
func domainToPRResponse(pr *domain.PullRequest) *PRResponse {
	resp := &PRResponse{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
	}

	if pr.CreatedAt != nil {
		resp.CreatedAt = pr.CreatedAt.Format(time.RFC3339)
	}
	if pr.MergedAt != nil {
		resp.MergedAt = pr.MergedAt.Format(time.RFC3339)
	}

	return resp
}
