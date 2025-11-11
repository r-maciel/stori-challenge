package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
