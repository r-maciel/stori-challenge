package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"stori-challenge/internal/domain"
	"stori-challenge/internal/ports/repositories"
)

type TransactionRepo struct {
	DB *sql.DB
}

var _ repositories.TransactionRepository = (*TransactionRepo)(nil)

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{DB: db}
}

func (r *TransactionRepo) ExistsByIDs(ctx context.Context, ids []int64) (map[int64]bool, error) {
	result := make(map[int64]bool)
	if len(ids) == 0 {
		return result, nil
	}
	// Build placeholders $1,$2,...
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := fmt.Sprintf(`SELECT id FROM transactions WHERE id IN (%s)`, strings.Join(ph, ","))
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	return result, rows.Err()
}

func (r *TransactionRepo) BulkInsert(ctx context.Context, txs []domain.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Batch multi-row insert to avoid too-large statements
	const batchSize = 500
	for i := 0; i < len(txs); i += batchSize {
		end := i + batchSize
		if end > len(txs) {
			end = len(txs)
		}
		if err := insertBatch(ctx, tx, txs[i:end]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func insertBatch(ctx context.Context, tx *sql.Tx, txs []domain.Transaction) error {
	var (
		sb   strings.Builder
		args []any
	)
	sb.WriteString("INSERT INTO transactions (id, user_id, amount, datetime, type) VALUES ")
	for i, t := range txs {
		if i > 0 {
			sb.WriteString(",")
		}
		// 5 placeholders per row
		base := i*5 + 1
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", base, base+1, base+2, base+3, base+4))
		args = append(args, t.ID, t.UserID, t.Amount.StringFixed(2), t.DateTime.UTC(), string(t.Type))
	}
	// Ensure timestamptz type correct by casting if needed
	stmt := sb.String()
	_, err := tx.ExecContext(ctx, stmt, args...)
	return err
}

func (r *TransactionRepo) UserHasAnyTransaction(ctx context.Context, userID int64) (bool, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT 1 FROM transactions WHERE user_id = $1 LIMIT 1`, userID)
	var one int
	if err := row.Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *TransactionRepo) GetUserBalanceSummary(ctx context.Context, userID int64, from, to time.Time) (balance decimal.Decimal, totalDebits decimal.Decimal, totalCredits decimal.Decimal, err error) {
	// Use COALESCE to avoid NULLs; scan as strings to ensure precision with decimal.
	const q = `
SELECT
	COALESCE(SUM(amount), 0)::text AS balance,
	COALESCE(SUM(CASE WHEN type = 'debit' THEN -amount ELSE 0 END), 0)::text AS total_debits,
	COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0)::text AS total_credits
FROM transactions
WHERE user_id = $1 AND datetime BETWEEN $2 AND $3`
	var balStr, debStr, credStr string
	if err = r.DB.QueryRowContext(ctx, q, userID, from.UTC(), to.UTC()).Scan(&balStr, &debStr, &credStr); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	balance, err = decimal.NewFromString(balStr)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	totalDebits, err = decimal.NewFromString(debStr)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	totalCredits, err = decimal.NewFromString(credStr)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	return balance, totalDebits, totalCredits, nil
}
