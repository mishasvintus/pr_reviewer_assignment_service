package repository

import "database/sql"

// DBTX is a common interface for *sql.DB and *sql.Tx.
// Both types implement the same methods for executing SQL queries.
type DBTX interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Compile-time check that *sql.DB and *sql.Tx implement DBTX.
var (
	_ DBTX = (*sql.DB)(nil)
	_ DBTX = (*sql.Tx)(nil)
)
