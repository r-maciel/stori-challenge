package repositories

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"stori-challenge/internal/domain"
)

type TransactionRepository interface {
	// ExistsByIDs returns a map of id -> true for any ids that already exist.
	ExistsByIDs(ctx context.Context, ids []int64) (map[int64]bool, error)
	// BulkInsert inserts all transactions in a single transaction (all-or-nothing).
	BulkInsert(ctx context.Context, txs []domain.Transaction) error
	// UserHasAnyTransaction returns true if the user has at least one transaction (any datetime).
	UserHasAnyTransaction(ctx context.Context, userID int64) (bool, error)
	// GetUserBalanceSummary returns the aggregated balance and totals within [from, to].
	// - balance: SUM(amount)
	// - totalDebits: SUM(-amount) for type = 'debit' (positive magnitude)
	// - totalCredits: SUM(amount) for type = 'credit'
	GetUserBalanceSummary(ctx context.Context, userID int64, from, to time.Time) (balance decimal.Decimal, totalDebits decimal.Decimal, totalCredits decimal.Decimal, err error)
}
