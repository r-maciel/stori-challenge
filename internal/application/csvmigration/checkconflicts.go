package csvmigration

import (
	"context"
	"strconv"

	"stori-challenge/internal/domain"
	"stori-challenge/internal/ports/services"
)

// checkConflicts asks the repository for existing ids.
func (s *csvMigrationService) checkConflicts(ctx context.Context, txs []domain.Transaction) (map[int64]bool, error) {
	ids := make([]int64, 0, len(txs))
	for _, tx := range txs {
		ids = append(ids, tx.ID)
	}
	return s.Repo.ExistsByIDs(ctx, ids)
}

// buildConflictErrors maps existing ids back to row numbers and values for error reporting.
func (s *csvMigrationService) buildConflictErrors(rows []ParsedRow, existing map[int64]bool) []services.RowError {
	var errs []services.RowError
	for _, pr := range rows {
		if pr.IDStr == "" {
			continue
		}
		if id, err := strconv.ParseInt(pr.IDStr, 10, 64); err == nil {
			if existing[id] {
				errs = append(errs, services.RowError{
					Row:     pr.RowNum,
					Field:   "id",
					Value:   pr.IDStr,
					Message: "id already exists in DB",
				})
			}
		}
	}
	return errs
}
