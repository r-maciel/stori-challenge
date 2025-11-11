package services

import (
	"context"
	"io"
)

// RowError represents a single validation/conflict detail for a CSV row in the migration use case.
type RowError struct {
	Row     int
	Field   string
	Value   string
	Message string
}

// MigrationService is the input port for the POST /migrate use case.
type MigrationService interface {
	// Process reads a CSV stream and returns:
	// - inserted: number of inserted rows on success
	// - items: row-level validation or conflict details
	// - err: typed error indicating class (BadRequest/Conflict/NotFound/Internal)
	Process(ctx context.Context, r io.Reader) (inserted int, items []RowError, err error)
}
