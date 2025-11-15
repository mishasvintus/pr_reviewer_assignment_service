package team

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// Create inserts a new team.
func Create(exec repository.DBTX, teamName string) error {
	query := `INSERT INTO teams (team_name) VALUES ($1)`
	_, err := exec.Exec(query, teamName)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

// Get retrieves a team with all its members.
func Get(exec repository.DBTX, teamName string) (*domain.Team, error) {
	query := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
	`
	rows, err := exec.Query(query, teamName)
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

	// If no members, check if team exists
	if len(members) == 0 {
		exists, err := Exists(exec, teamName)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, sql.ErrNoRows
		}
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

// Exists checks if a team exists.
func Exists(exec repository.DBTX, teamName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	err := exec.QueryRow(query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}
