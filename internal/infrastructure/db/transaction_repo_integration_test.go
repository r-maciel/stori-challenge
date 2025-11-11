package db

import (
	"context"
	"testing"
	"time"

	"stori-challenge/internal/domain"
	testinfra "stori-challenge/internal/shared/test"

	"github.com/shopspring/decimal"
)

func TestIntegration_BulkInsert_And_ExistsByIDs(t *testing.T) {
	db, err := testinfra.OpenTestDB(t.Name())
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	repo := NewTransactionRepo(db)

	// clean slate: ensure ids do not exist
	ctx := context.Background()
	ids := []int64{1001, 1002}
	exists, err := repo.ExistsByIDs(ctx, ids)
	if err != nil {
		t.Fatalf("exists pre-check: %v", err)
	}
	if len(exists) != 0 {
		t.Fatalf("expected none to exist initially, got %v", exists)
	}

	// insert
	txs := []domain.Transaction{
		{ID: 1001, UserID: 10, Amount: decimal.NewFromFloat(1.23), DateTime: time.Unix(0, 0).UTC(), Type: domain.TransactionTypeCredit},
		{ID: 1002, UserID: 20, Amount: decimal.NewFromFloat(4.56), DateTime: time.Unix(10, 0).UTC(), Type: domain.TransactionTypeCredit},
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("bulkinsert: %v", err)
	}
	// verify
	exists, err = repo.ExistsByIDs(ctx, ids)
	if err != nil {
		t.Fatalf("exists after insert: %v", err)
	}
	if len(exists) != 2 || !exists[1001] || !exists[1002] {
		t.Fatalf("expected both to exist, got %v", exists)
	}
}

func TestIntegration_BulkInsert_Duplicate_FailsAndRollsBack(t *testing.T) {
	db, err := testinfra.OpenTestDB(t.Name())
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	repo := NewTransactionRepo(db)
	ctx := context.Background()

	txs := []domain.Transaction{
		{ID: 2001, UserID: 10, Amount: decimal.NewFromFloat(1), DateTime: time.Unix(0, 0).UTC(), Type: domain.TransactionTypeCredit},
		{ID: 2001, UserID: 11, Amount: decimal.NewFromFloat(2), DateTime: time.Unix(10, 0).UTC(), Type: domain.TransactionTypeCredit}, // duplicate id
	}
	if err := repo.BulkInsert(ctx, txs); err == nil {
		t.Fatalf("expected duplicate to fail")
	}
	// verify rollback: none should exist
	exists, err := repo.ExistsByIDs(ctx, []int64{2001})
	if err != nil {
		t.Fatalf("exists after failure: %v", err)
	}
	if len(exists) != 0 {
		t.Fatalf("expected no rows after rollback, got %v", exists)
	}
}

func TestIntegration_BulkInsert_BatchOver500_Succeeds(t *testing.T) {
	db, err := testinfra.OpenTestDB(t.Name())
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	repo := NewTransactionRepo(db)
	ctx := context.Background()

	count := 501
	txs := make([]domain.Transaction, 0, count)
	ids := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		id := int64(3000 + i)
		ids = append(ids, id)
		txs = append(txs, domain.Transaction{
			ID:       id,
			UserID:   999,
			Amount:   decimal.NewFromFloat(1.11),
			DateTime: time.Unix(int64(i), 0).UTC(),
			Type:     domain.TransactionTypeCredit,
		})
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("bulkinsert large: %v", err)
	}
	exists, err := repo.ExistsByIDs(ctx, ids)
	if err != nil {
		t.Fatalf("exists after large insert: %v", err)
	}
	if len(exists) != count {
		t.Fatalf("expected %d existing, got %d", count, len(exists))
	}
}
