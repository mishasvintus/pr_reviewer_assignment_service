package domain

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// PRStatus represents the status of a pull request.
type PRStatus string

// PR status constants.
const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

// NewPRStatus creates a new PRStatus with validation.
// Returns an error if the status is invalid.
func NewPRStatus(s string) (PRStatus, error) {
	status := PRStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid PR status: %s (must be one of: %s, %s)", s, StatusOpen, StatusMerged)
	}
	return status, nil
}

// IsValid checks if the status is valid.
func (s PRStatus) IsValid() bool {
	return s == StatusOpen || s == StatusMerged
}

// Scan implements sql.Scanner interface for automatic validation when reading from database.
func (s *PRStatus) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("PRStatus cannot be NULL")
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan %T into PRStatus", value)
	}

	status, err := NewPRStatus(str)
	if err != nil {
		return err
	}
	*s = status
	return nil
}

// Value implements driver.Valuer interface for writing to database.
func (s PRStatus) Value() (driver.Value, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid PRStatus value: %s", s)
	}
	return string(s), nil
}

// PullRequest represents a pull request with assigned reviewers.
type PullRequest struct {
	PullRequestID        string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName      string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID             string     `json:"author_id" db:"author_id"`
	TeamName             string     `json:"team_name" db:"team_name"`
	Status               PRStatus   `json:"status" db:"status"`
	AssignedReviewersIDs []string   `json:"assigned_reviewers"`
	CreatedAt            *time.Time `json:"createdAt,omitempty" db:"created_at"`
	MergedAt             *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

// PullRequestShort is a lightweight version of PullRequest for lists.
type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	TeamName        string   `json:"team_name"`
	Status          PRStatus `json:"status"`
}
