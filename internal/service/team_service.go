package service

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/pr"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/team"
	"github.com/mishasvintus/avito_backend_internship/internal/repository/user"
)

// TeamService handles team business logic.
type TeamService struct {
	db        *sql.DB
	prService *PRService
}

// NewTeamService creates a new team service.
func NewTeamService(db *sql.DB, prService *PRService) *TeamService {
	return &TeamService{db: db, prService: prService}
}

// CreateTeam creates a new team with members in a single transaction.
func (s *TeamService) CreateTeam(teamName string, members []domain.TeamMember) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

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

// DeactivateTeam deactivates all users in a team and reassigns open PRs.
func (s *TeamService) DeactivateTeam(teamName string) error {
	// Check if team exists
	_, err := team.Get(s.db, teamName)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTeamNotFound
		}
		return fmt.Errorf("failed to check team: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Deactivate all team users
	if err := team.DeactivateAll(tx, teamName); err != nil {
		return fmt.Errorf("failed to deactivate team: %w", err)
	}

	// 2. Find open PRs that have reviewers from this team 
	prReviewers, err := pr.GetOpenPRsWithReviewersFromTeam(tx, teamName)
	if err != nil {
		return fmt.Errorf("failed to get open PRs: %w", err)
	}

	// 3. For each PR: remove reviewers from the team, then replenish from PR's team if needed
	for prID, reviewerIDs := range prReviewers {
		for _, reviewerID := range reviewerIDs {
			if err := pr.DeleteReviewer(tx, prID, reviewerID); err != nil {
				return fmt.Errorf("failed to delete reviewer: %w", err)
			}
		}

		pullRequest, err := pr.Get(tx, prID)
		if err != nil {
			return fmt.Errorf("failed to get PR: %w", err)
		}
		if pullRequest.TeamName == teamName {
			continue
		}
		if err := s.prService.ReplenishReviewers(tx, prID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
