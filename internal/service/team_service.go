package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/team"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
)

// TeamService handles team business logic.
type TeamService struct {
	db *sql.DB
}

// NewTeamService creates a new team service.
func NewTeamService(db *sql.DB) *TeamService {
	return &TeamService{db: db}
}

// CreateTeam creates a new team with members in a single transaction.
func (s *TeamService) CreateTeam(teamName string, members []domain.TeamMember) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if team already exists
	exists, err := team.Exists(tx, teamName)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return ErrTeamExists
	}

	// Create team
	if err := team.Create(tx, teamName); err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	// Process each user: create if not exists, update if exists
	for _, member := range members {
		u := domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: teamName,
			IsActive: member.IsActive,
		}

		// Check if user exists
		existingUser, err := user.Get(tx, member.UserID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check user existence: %w", err)
		}

		if existingUser == nil {
			if err := user.Create(tx, &u); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			if err := user.Update(tx, &u); err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTeam retrieves a team with all its members.
func (s *TeamService) GetTeam(teamName string) (*domain.Team, error) {
	t, err := team.Get(s.db, teamName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return t, nil
}
