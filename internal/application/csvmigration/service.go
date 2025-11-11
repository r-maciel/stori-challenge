package csvmigration

import (
	"context"
	"io"
	"time"

	"stori-challenge/internal/ports/repositories"
	"stori-challenge/internal/ports/services"
	"stori-challenge/internal/shared"
)

// expectedHeaders defines the strict order of required columns.
var expectedHeaders = []string{"id", "user_id", "amount", "datetime"}

// ParsedRow holds raw string values and metadata for error reporting.
type ParsedRow struct {
	RowNum      int
	IDStr       string
	UserIDStr   string
	AmountStr   string
	DatetimeStr string
	Cols        int
}

// csvMigrationService implements services.MigrationService for CSV inputs.
type csvMigrationService struct {
	Repo    repositories.TransactionRepository
	NowFunc func() time.Time
}

// NewCsvMigrationService constructs a CSV migration service.
func NewCsvMigrationService(repo repositories.TransactionRepository) services.MigrationService {
	return &csvMigrationService{
		Repo:    repo,
		NowFunc: func() time.Time { return time.Now().UTC() },
	}
}

// Ensure interface compliance.
var _ services.MigrationService = (*csvMigrationService)(nil)

// Process runs intake (caller validates file), parse+validate (single pass), conflict check, and bulk insert.
func (s *csvMigrationService) Process(ctx context.Context, r io.Reader) (int, []services.RowError, error) {
	txs, rows, vErrs, parseErr := s.readAndValidate(r)
	if parseErr != nil {
		// Header or CSV malformed → treat as validation error
		return 0, []services.RowError{{
			Row:     0,
			Field:   "file",
			Value:   "",
			Message: "invalid or missing header",
		}}, shared.NewBadRequest("validation_error", "validation failed", parseErr)
	}

	if len(vErrs) > 0 {
		return 0, vErrs, shared.NewBadRequest("validation_error", "validation failed", nil)
	}
	if len(txs) == 0 {
		return 0, []services.RowError{{
			Row:     0,
			Field:   "file",
			Value:   "",
			Message: "CSV contains no data rows",
		}}, shared.NewBadRequest("validation_error", "validation failed", nil)
	}

	existing, err := s.checkConflicts(ctx, txs)
	if err != nil {
		return 0, nil, shared.NewInternal("db_failure", "database error", err)
	}
	if len(existing) > 0 {
		cErrs := s.buildConflictErrors(rows, existing)
		return 0, cErrs, shared.NewConflict("duplicate_id", "conflict", nil)
	}

	if err := s.insertAll(ctx, txs); err != nil {
		return 0, nil, shared.NewInternal("db_failure", "database error", err)
	}
	return len(txs), nil, nil
}
