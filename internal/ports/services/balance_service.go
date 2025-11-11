package services

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// BalanceService exposes user balance aggregation operations.
type BalanceService interface {
	// GetBalance returns aggregated amounts within [from, to] for the given user.
	// Returns not found if the user has no transactions at all.
	GetBalance(ctx context.Context, userID int64, from, to time.Time) (balance decimal.Decimal, totalDebits decimal.Decimal, totalCredits decimal.Decimal, err error)
}


