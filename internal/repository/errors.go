package repository

import (
	"errors"

	"github.com/lib/pq"
)

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
// PostgreSQL error code 23505 = unique_violation.
func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// IsForeignKeyViolation checks if the error is a PostgreSQL foreign key violation.
// PostgreSQL error code 23503 = foreign_key_violation.
func IsForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23503"
	}
	return false
}
