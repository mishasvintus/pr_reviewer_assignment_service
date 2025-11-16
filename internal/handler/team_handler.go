package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mishasvintus/avito_backend_internship/internal/service"
)

// TeamHandler handles team-related HTTP requests.
type TeamHandler struct {
	teamService *service.TeamService
}

// NewTeamHandler creates a new team handler.
func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// AddTeam handles POST /team/add.
func (h *TeamHandler) AddTeam(c *gin.Context) {
	var req AddTeamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	err := h.teamService.CreateTeam(req.TeamName, req.Members)
	if err != nil {
		if errors.Is(err, service.ErrTeamExists) {
			Error(c, ErrorTeamExists, "team_name already exists", http.StatusBadRequest)
			return
		}
		InternalError(c, err.Error())
		return
	}

	team, err := h.teamService.GetTeam(req.TeamName)
	if err != nil {
		InternalError(c, "failed to retrieve created team")
		return
	}

	members := make([]TeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Team: &TeamResponse{
			TeamName: team.TeamName,
			Members:  members,
		},
	})
}

// GetTeam handles GET /team/get.
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		BadRequest(c, "team_name parameter is required")
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, service.ErrTeamNotFound) {
			NotFound(c, "team not found")
			return
		}
		InternalError(c, err.Error())
		return
	}

	members := make([]TeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	c.JSON(http.StatusOK, TeamResponse{
		TeamName: team.TeamName,
		Members:  members,
	})
}
