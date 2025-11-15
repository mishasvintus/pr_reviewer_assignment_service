package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// TeamService handles team business logic.
type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

// NewTeamService creates a new team service.
func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeam creates a new team with members in a single transaction.
// Returns error if team or any user_id already exists (relies on database constraints).
func (s *TeamService) CreateTeam(teamName string, members []domain.TeamMember) error {
	users := make([]domain.User, len(members))
	for i, member := range members {
		users[i] = domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: teamName,
			IsActive: member.IsActive,
		}
	}

	err := s.teamRepo.CreateWithMembers(teamName, users)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return fmt.Errorf("team or user already exists")
		}
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

// GetTeam retrieves a team with all its members.
func (s *TeamService) GetTeam(teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.Get(teamName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}
