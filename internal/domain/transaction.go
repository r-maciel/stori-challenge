package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID       int64
	UserID   int64
	Amount   decimal.Decimal
	DateTime time.Time
}


