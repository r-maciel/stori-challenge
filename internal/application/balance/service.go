package balance

import (
	"context"
	"time"

	"stori-challenge/internal/ports/repositories"
	"stori-challenge/internal/ports/services"
	"stori-challenge/internal/shared"

	"github.com/shopspring/decimal"
)

type balanceService struct {
	Repo repositories.TransactionRepository
}

// Ensure interface compliance
var _ services.BalanceService = (*balanceService)(nil)

func NewBalanceService(repo repositories.TransactionRepository) services.BalanceService {
	return &balanceService{Repo: repo}
}

func (s *balanceService) GetBalance(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	// First, ensure the user has any transactions at all
	hasAny, err := s.Repo.UserHasAnyTransaction(ctx, userID)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, shared.NewInternal("db_failure", "database error", err)
	}
	if !hasAny {
		return decimal.Zero, decimal.Zero, decimal.Zero, shared.NewNotFound("user_transactions_not_found", "user has no transactions", nil)
	}
	// Aggregate within the provided window
	bal, deb, cred, err := s.Repo.GetUserBalanceSummary(ctx, userID, from, to)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, shared.NewInternal("db_failure", "database error", err)
	}
	return bal, deb, cred, nil
}
