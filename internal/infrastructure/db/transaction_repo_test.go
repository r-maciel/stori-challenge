package db

import (
	"context"
	"regexp"
	"testing"
	"time"

	"stori-challenge/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

func TestExistsByIDs_EmptyInput_ReturnsEmpty(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()

	repo := NewTransactionRepo(sqlDB)
	got, err := repo.ExistsByIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}
}

func TestExistsByIDs_ReturnsPresentMap(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()

	repo := NewTransactionRepo(sqlDB)

	ids := []int64{10, 20, 30}
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(20)).AddRow(int64(30))
	queryRe := regexp.MustCompile(`SELECT id FROM transactions WHERE id IN \(\$1,\$2,\$3\)`)
	mock.ExpectQuery(queryRe.String()).WithArgs(ids[0], ids[1], ids[2]).WillReturnRows(rows)

	got, err := repo.ExistsByIDs(context.Background(), ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || !got[20] || !got[30] {
		t.Fatalf("unexpected result map: %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExistsByIDs_QueryError_Propagates(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	ids := []int64{1}
	queryRe := regexp.MustCompile(`SELECT id FROM transactions WHERE id IN \(\$1\)`)
	mock.ExpectQuery(queryRe.String()).WithArgs(ids[0]).WillReturnError(assertErr)

	_, err = repo.ExistsByIDs(context.Background(), ids)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestBulkInsert_Empty_Noop(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)
	if err := repo.BulkInsert(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBulkInsert_TwoRows_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	t1 := domain.Transaction{
		ID:       1,
		UserID:   100,
		Amount:   decimal.NewFromFloat(12.34),
		DateTime: time.Date(2023, 10, 1, 15, 4, 5, 0, time.UTC),
		Type:     domain.TransactionTypeCredit,
	}
	t2 := domain.Transaction{
		ID:       2,
		UserID:   200,
		Amount:   decimal.NewFromFloat(56.78),
		DateTime: time.Date(2023, 10, 2, 11, 0, 0, 0, time.UTC),
		Type:     domain.TransactionTypeCredit,
	}

	mock.ExpectBegin()
	// Match the INSERT statement shape; placeholders grow with rows.
	stmtRe := regexp.MustCompile(`INSERT INTO transactions \(id, user_id, amount, datetime, type\) VALUES \(\$1,\$2,\$3,\$4,\$5\),\(\$6,\$7,\$8,\$9,\$10\)`)
	mock.ExpectExec(stmtRe.String()).
		WithArgs(
			int64(1), int64(100), "12.34", t1.DateTime.UTC(), "credit",
			int64(2), int64(200), "56.78", t2.DateTime.UTC(), "credit",
		).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	if err := repo.BulkInsert(context.Background(), []domain.Transaction{t1, t2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBulkInsert_BeginError_Propagates(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	mock.ExpectBegin().WillReturnError(assertErr)
	err = repo.BulkInsert(context.Background(), []domain.Transaction{{ID: 1, UserID: 1, Amount: decimal.NewFromInt(1), DateTime: time.Unix(0, 0).UTC()}})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestBulkInsert_ExecError_PropagatesAndRollbacks(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	t1 := domain.Transaction{ID: 1, UserID: 100, Amount: decimal.NewFromInt(1), DateTime: time.Unix(0, 0).UTC(), Type: domain.TransactionTypeCredit}

	mock.ExpectBegin()
	stmtRe := regexp.MustCompile(`INSERT INTO transactions \(id, user_id, amount, datetime, type\) VALUES \(\$1,\$2,\$3,\$4,\$5\)`)
	mock.ExpectExec(stmtRe.String()).
		WithArgs(int64(1), int64(100), "1.00", t1.DateTime.UTC(), "credit").
		WillReturnError(assertErr)
	mock.ExpectRollback()

	err = repo.BulkInsert(context.Background(), []domain.Transaction{t1})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBulkInsert_CommitError_Propagates(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	t1 := domain.Transaction{ID: 1, UserID: 100, Amount: decimal.NewFromInt(1), DateTime: time.Unix(0, 0).UTC(), Type: domain.TransactionTypeCredit}

	mock.ExpectBegin()
	stmtRe := regexp.MustCompile(`INSERT INTO transactions \(id, user_id, amount, datetime, type\) VALUES \(\$1,\$2,\$3,\$4,\$5\)`)
	mock.ExpectExec(stmtRe.String()).
		WithArgs(int64(1), int64(100), "1.00", t1.DateTime.UTC(), "credit").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(assertErr)

	err = repo.BulkInsert(context.Background(), []domain.Transaction{t1})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// assertErr is a sentinel error used in expectations.
type testError string

func (e testError) Error() string { return string(e) }

const assertErr = testError("assert-err")

func TestUserHasAnyTransaction_FalseWhenNoRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	userID := int64(777)
	queryRe := regexp.MustCompile(`SELECT 1 FROM transactions WHERE user_id = \$1 LIMIT 1`)
	mock.ExpectQuery(queryRe.String()).WithArgs(userID).WillReturnRows(sqlmock.NewRows([]string{"one"}))

	got, err := repo.UserHasAnyTransaction(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatalf("expected false, got true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserHasAnyTransaction_TrueWhenRowExists(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	userID := int64(888)
	queryRe := regexp.MustCompile(`SELECT 1 FROM transactions WHERE user_id = \$1 LIMIT 1`)
	mock.ExpectQuery(queryRe.String()).WithArgs(userID).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))

	got, err := repo.UserHasAnyTransaction(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatalf("expected true, got false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetUserBalanceSummary_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	userID := int64(123)
	from := time.Unix(0, 0).UTC()
	to := time.Unix(1000, 0).UTC()
	stmtRe := regexp.MustCompile(`SELECT\s+COALESCE\(SUM\(amount\), 0\)::text AS balance,\s+COALESCE\(SUM\(CASE WHEN type = 'debit' THEN -amount ELSE 0 END\), 0\)::text AS total_debits,\s+COALESCE\(SUM\(CASE WHEN type = 'credit' THEN amount ELSE 0 END\), 0\)::text AS total_credits\s+FROM transactions\s+WHERE user_id = \$1 AND datetime BETWEEN \$2 AND \$3`)
	rows := sqlmock.NewRows([]string{"balance", "total_debits", "total_credits"}).AddRow("5.21", "10.00", "15.21")
	mock.ExpectQuery(stmtRe.String()).WithArgs(userID, from, to).WillReturnRows(rows)

	bal, deb, cred, err := repo.GetUserBalanceSummary(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bal.StringFixed(2) != "5.21" || deb.StringFixed(2) != "10.00" || cred.StringFixed(2) != "15.21" {
		t.Fatalf("unexpected values: %s %s %s", bal.StringFixed(2), deb.StringFixed(2), cred.StringFixed(2))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetUserBalanceSummary_QueryError_Propagates(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()
	repo := NewTransactionRepo(sqlDB)

	userID := int64(123)
	from := time.Unix(0, 0).UTC()
	to := time.Unix(1000, 0).UTC()
	stmtRe := regexp.MustCompile(`SELECT\s+COALESCE\(SUM\(amount\), 0\)::text AS balance,\s+COALESCE\(SUM\(CASE WHEN type = 'debit' THEN -amount ELSE 0 END\), 0\)::text AS total_debits,\s+COALESCE\(SUM\(CASE WHEN type = 'credit' THEN amount ELSE 0 END\), 0\)::text AS total_credits\s+FROM transactions\s+WHERE user_id = \$1 AND datetime BETWEEN \$2 AND \$3`)
	mock.ExpectQuery(stmtRe.String()).WithArgs(userID, from, to).WillReturnError(assertErr)

	_, _, _, err = repo.GetUserBalanceSummary(context.Background(), userID, from, to)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
