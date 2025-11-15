package repository

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
// PostgreSQL error code 23505 = unique_violation.
func IsUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

// IsForeignKeyViolation checks if the error is a PostgreSQL foreign key violation.
// PostgreSQL error code 23503 = foreign_key_violation.
func IsForeignKeyViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503"
	}
	return false
}

// ErrInactiveReviewer is returned when trying to assign an inactive user as a reviewer.
type ErrInactiveReviewer struct {
	UserID string
}

func (e *ErrInactiveReviewer) Error() string {
	return fmt.Sprintf("reviewer %s is not active", e.UserID)
}

// IsInactiveReviewer checks if the error is an inactive reviewer error.
func IsInactiveReviewer(err error) bool {
	var inactiveErr *ErrInactiveReviewer
	return errors.As(err, &inactiveErr)
}
