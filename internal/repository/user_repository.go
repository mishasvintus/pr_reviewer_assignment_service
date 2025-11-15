package repository

import (
	"database/sql"
	"fmt"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
)

// UserRepository handles user database operations.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(query, user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// CreateBatch creates multiple users in a single transaction.
func (r *UserRepository) CreateBatch(users []domain.User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, user := range users {
		_, err := stmt.Exec(user.UserID, user.Username, user.TeamName, user.IsActive)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", user.UserID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Get retrieves a user by ID.
func (r *UserRepository) Get(userID string) (*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`
	var user domain.User
	err := r.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// SetIsActive updates the is_active status of a user.
func (r *UserRepository) SetIsActive(userID string, isActive bool) error {
	query := `UPDATE users SET is_active = $1 WHERE user_id = $2`
	result, err := r.db.Exec(query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
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

// Exists checks if a user exists.
func (r *UserRepository) Exists(userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// GetActiveTeammates returns all active users from the same team, excluding the given user.
func (r *UserRepository) GetActiveTeammates(userID string) ([]domain.User, error) {
	query := `
		SELECT u2.user_id, u2.username, u2.team_name, u2.is_active
		FROM users u1
		JOIN users u2 ON u1.team_name = u2.team_name
		WHERE u1.user_id = $1 
		  AND u2.user_id != $1
		  AND u2.is_active = true
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active teammates: %w", err)
	}
	defer rows.Close()

	var teammates []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan teammate: %w", err)
		}
		teammates = append(teammates, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return teammates, nil
}
