package repository

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
)

// TeamRepository handles team database operations.
type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository creates a new team repository.
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team in the database.
func (r *TeamRepository) Create(teamName string) error {
	query := `INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING`
	_, err := r.db.Exec(query, teamName)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

// Get retrieves a team with all its members.
func (r *TeamRepository) Get(teamName string) (*domain.Team, error) {
	// Get team members
	query := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
	`
	rows, err := r.db.Query(query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// If we have members, team exists
	if len(members) > 0 {
		return &domain.Team{
			TeamName: teamName,
			Members:  members,
		}, nil
	}

	// No members - check if team exists
	exists, err := r.Exists(teamName)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, sql.ErrNoRows
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

// Exists checks if a team exists.
func (r *TeamRepository) Exists(teamName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	err := r.db.QueryRow(query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}
