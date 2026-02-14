package user

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
	"github.com/mishasvintus/avito_backend_internship/internal/repository"
)

// Create inserts a new user.
func Create(exec repository.DBTX, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`
	_, err := exec.Exec(query, user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// Get retrieves a user by ID.
func Get(exec repository.DBTX, userID string) (*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`
	var u domain.User
	err := exec.QueryRow(query, userID).Scan(
		&u.UserID,
		&u.Username,
		&u.TeamName,
		&u.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

// Update updates user's team_name, username, and is_active.
func Update(exec repository.DBTX, user *domain.User) error {
	query := `
		UPDATE users 
		SET username = $1, team_name = $2, is_active = $3
		WHERE user_id = $4
	`
	result, err := exec.Exec(query, user.Username, user.TeamName, user.IsActive, user.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SetIsActive updates the is_active status and returns the updated user.
func SetIsActive(exec repository.DBTX, userID string, isActive bool) (*domain.User, error) {
	query := `
		UPDATE users 
		SET is_active = $1 
		WHERE user_id = $2 
		RETURNING user_id, username, team_name, is_active
	`
	var u domain.User
	err := exec.QueryRow(query, isActive, userID).Scan(
		&u.UserID,
		&u.Username,
		&u.TeamName,
		&u.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return &u, nil
}

// GetActiveTeammates returns all active users from the same team, excluding the given user.
func GetActiveTeammates(exec repository.DBTX, userID string) ([]domain.User, error) {
	query := `
		SELECT u2.user_id, u2.username, u2.team_name, u2.is_active
		FROM users u1
		JOIN users u2 ON u1.team_name = u2.team_name
		WHERE u1.user_id = $1 
		  AND u2.user_id != $1
		  AND u2.is_active = true
	`
	rows, err := exec.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active teammates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var teammates []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan teammate: %w", err)
		}
		teammates = append(teammates, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return teammates, nil
}

// GetActiveByTeam returns all active users in the given team.
func GetActiveByTeam(exec repository.DBTX, teamName string) ([]domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = true
	`
	rows, err := exec.Query(query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users by team: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}
