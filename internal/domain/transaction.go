package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransactionType represents the kind of transaction based on amount sign.
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"
)

// DetermineTransactionType returns debit if amount < 0, otherwise credit.
func DetermineTransactionType(amount decimal.Decimal) TransactionType {
	if amount.IsNegative() {
		return TransactionTypeDebit
	}
	return TransactionTypeCredit
}

type Transaction struct {
	ID       int64
	UserID   int64
	Amount   decimal.Decimal
	DateTime time.Time
	Type     TransactionType
}


