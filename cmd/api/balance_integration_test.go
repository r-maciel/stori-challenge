package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stori-challenge/internal/domain"
	infradb "stori-challenge/internal/infrastructure/db"
	"stori-challenge/internal/infrastructure/http/responses"

	"github.com/shopspring/decimal"
)

func TestBalanceIntegration_Success200_NoFilters(t *testing.T) {
	router, db := newTestRouter(t)
	repo := infradb.NewTransactionRepo(db)
	ctx := context.Background()
	userID := int64(501)
	now := time.Now().UTC()
	// Seed transactions for user 501
	txs := []domain.Transaction{
		{ID: 90001, UserID: userID, Amount: decimal.NewFromFloat(10.00), DateTime: now.Add(-48 * time.Hour), Type: domain.TransactionTypeCredit},
		{ID: 90002, UserID: userID, Amount: decimal.NewFromFloat(-3.50), DateTime: now.Add(-24 * time.Hour), Type: domain.TransactionTypeDebit},
		{ID: 90003, UserID: userID, Amount: decimal.NewFromFloat(5.71), DateTime: now.Add(-1 * time.Hour), Type: domain.TransactionTypeCredit},
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("seed insert: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/501/balance", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d; body=%s", w.Code, w.Body.String())
	}
	var ok responses.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Expected:
	// balance = 10.00 - 3.50 + 5.71 = 12.21
	// total_debits = 3.50 (positive magnitude)
	// total_credits = 15.71
	if ok.Balance != 12.21 || ok.TotalDebits != 3.50 || ok.TotalCredits != 15.71 {
		t.Fatalf("payload mismatch: %+v", ok)
	}
}

func TestBalanceIntegration_NotFound404_WhenUserHasNoTransactions(t *testing.T) {
	router, _ := newTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/999999/balance", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status: want 404 got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestBalanceIntegration_BadRequest400_InvalidDate(t *testing.T) {
	router, _ := newTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/1/balance?from=2024-01-01T00:00:00", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: want 400 got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestBalanceIntegration_Success200_OnlyFrom(t *testing.T) {
	router, db := newTestRouter(t)
	repo := infradb.NewTransactionRepo(db)
	ctx := context.Background()
	userID := int64(601)
	now := time.Now().UTC()
	from := now.Add(-36 * time.Hour).Format(time.RFC3339)
	// Seed: one before from (excluded), two after from (included)
	txs := []domain.Transaction{
		{ID: 91001, UserID: userID, Amount: decimal.NewFromFloat(5.00), DateTime: now.Add(-72 * time.Hour), Type: domain.TransactionTypeCredit}, // excluded
		{ID: 91002, UserID: userID, Amount: decimal.NewFromFloat(2.00), DateTime: now.Add(-24 * time.Hour), Type: domain.TransactionTypeCredit}, // included
		{ID: 91003, UserID: userID, Amount: decimal.NewFromFloat(-1.50), DateTime: now.Add(-1 * time.Hour), Type: domain.TransactionTypeDebit},  // included
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("seed insert: %v", err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/601/balance?from="+from, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d; body=%s", w.Code, w.Body.String())
	}
	var ok responses.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Included: +2.00 and -1.50 ⇒ balance 0.50, debits 1.50, credits 2.00
	if ok.Balance != 0.50 || ok.TotalDebits != 1.50 || ok.TotalCredits != 2.00 {
		t.Fatalf("payload mismatch: %+v", ok)
	}
}

func TestBalanceIntegration_Success200_OnlyTo(t *testing.T) {
	router, db := newTestRouter(t)
	repo := infradb.NewTransactionRepo(db)
	ctx := context.Background()
	userID := int64(602)
	now := time.Now().UTC()
	// to acts as lower bound per reglas; upper = now
	toLower := now.Add(-24 * time.Hour).Format(time.RFC3339)
	// Seed: one before lower (excluded), two after lower (included)
	txs := []domain.Transaction{
		{ID: 92001, UserID: userID, Amount: decimal.NewFromFloat(7.00), DateTime: now.Add(-48 * time.Hour), Type: domain.TransactionTypeCredit}, // excluded
		{ID: 92002, UserID: userID, Amount: decimal.NewFromFloat(3.00), DateTime: now.Add(-12 * time.Hour), Type: domain.TransactionTypeCredit}, // included
		{ID: 92003, UserID: userID, Amount: decimal.NewFromFloat(-2.00), DateTime: now.Add(-2 * time.Hour), Type: domain.TransactionTypeDebit},  // included
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("seed insert: %v", err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/602/balance?to="+toLower, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d; body=%s", w.Code, w.Body.String())
	}
	var ok responses.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Included: +3.00 and -2.00 ⇒ balance 1.00, debits 2.00, credits 3.00
	if ok.Balance != 1.00 || ok.TotalDebits != 2.00 || ok.TotalCredits != 3.00 {
		t.Fatalf("payload mismatch: %+v", ok)
	}
}

func TestBalanceIntegration_Success200_ReversedBounds(t *testing.T) {
	router, db := newTestRouter(t)
	repo := infradb.NewTransactionRepo(db)
	ctx := context.Background()
	userID := int64(603)
	now := time.Now().UTC()
	lower := now.Add(-30 * time.Hour)
	upper := now.Add(-6 * time.Hour)
	// Provide reversed: from=upper, to=lower
	// Seed: inside window two rows
	txs := []domain.Transaction{
		{ID: 93001, UserID: userID, Amount: decimal.NewFromFloat(4.00), DateTime: now.Add(-20 * time.Hour), Type: domain.TransactionTypeCredit}, // in-range
		{ID: 93002, UserID: userID, Amount: decimal.NewFromFloat(-1.00), DateTime: now.Add(-10 * time.Hour), Type: domain.TransactionTypeDebit}, // in-range
		{ID: 93003, UserID: userID, Amount: decimal.NewFromFloat(9.00), DateTime: now.Add(-40 * time.Hour), Type: domain.TransactionTypeCredit}, // out-of-range
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("seed insert: %v", err)
	}
	w := httptest.NewRecorder()
	path := "/v1/users/603/balance?from=" + upper.Format(time.RFC3339) + "&to=" + lower.Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d; body=%s", w.Code, w.Body.String())
	}
	var ok responses.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Included: +4.00 and -1.00 ⇒ balance 3.00, debits 1.00, credits 4.00
	if ok.Balance != 3.00 || ok.TotalDebits != 1.00 || ok.TotalCredits != 4.00 {
		t.Fatalf("payload mismatch: %+v", ok)
	}
}

func TestBalanceIntegration_BadRequest400_UpperAfterNow(t *testing.T) {
	router, _ := newTestRouter(t)
	now := time.Now().UTC()
	future := now.Add(48 * time.Hour).Format(time.RFC3339)
	past := now.Add(-24 * time.Hour).Format(time.RFC3339)
	// Provide a future bound among the two; max > now ⇒ 400
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/1/balance?from="+past+"&to="+future, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: want 400 got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestBalanceIntegration_Success200_BothParamsOrdered(t *testing.T) {
	router, db := newTestRouter(t)
	repo := infradb.NewTransactionRepo(db)
	ctx := context.Background()
	userID := int64(604)
	now := time.Now().UTC()
	from := now.Add(-36 * time.Hour)
	to := now.Add(-6 * time.Hour)
	// Seed: two in-range, two out-of-range
	txs := []domain.Transaction{
		{ID: 94001, UserID: userID, Amount: decimal.NewFromFloat(8.00), DateTime: now.Add(-72 * time.Hour), Type: domain.TransactionTypeCredit}, // out (before)
		{ID: 94002, UserID: userID, Amount: decimal.NewFromFloat(3.00), DateTime: now.Add(-30 * time.Hour), Type: domain.TransactionTypeCredit}, // in
		{ID: 94003, UserID: userID, Amount: decimal.NewFromFloat(-1.25), DateTime: now.Add(-10 * time.Hour), Type: domain.TransactionTypeDebit}, // in
		{ID: 94004, UserID: userID, Amount: decimal.NewFromFloat(9.00), DateTime: now.Add(-1 * time.Hour), Type: domain.TransactionTypeCredit},  // out (after)
	}
	if err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("seed insert: %v", err)
	}
	w := httptest.NewRecorder()
	path := "/v1/users/604/balance?from=" + from.Format(time.RFC3339) + "&to=" + to.Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d; body=%s", w.Code, w.Body.String())
	}
	var ok responses.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Included: +3.00 and -1.25 ⇒ balance 1.75, debits 1.25, credits 3.00
	if ok.Balance != 1.75 || ok.TotalDebits != 1.25 || ok.TotalCredits != 3.00 {
		t.Fatalf("payload mismatch: %+v", ok)
	}
}
