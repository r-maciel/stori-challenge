package csvmigration

import (
	"context"

	"stori-challenge/internal/domain"
)

// insertAll persists all transactions in a single transaction.
func (s *csvMigrationService) insertAll(ctx context.Context, txs []domain.Transaction) error {
	return s.Repo.BulkInsert(ctx, txs)
}
