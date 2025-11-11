package repositories

import (
	"context"

	"stori-challenge/internal/domain"
)

type TransactionRepository interface {
	// ExistsByIDs returns a map of id -> true for any ids that already exist.
	ExistsByIDs(ctx context.Context, ids []int64) (map[int64]bool, error)
	// BulkInsert inserts all transactions in a single transaction (all-or-nothing).
	BulkInsert(ctx context.Context, txs []domain.Transaction) error
}
