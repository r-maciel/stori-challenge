package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stori-challenge/internal/infrastructure/http/responses"
	"stori-challenge/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type mockBalanceService struct {
	GetBalanceFn func(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error)
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	return m.GetBalanceFn(ctx, userID, from, to)
}

func TestGetBalance_InvalidUserID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "user_id", Value: "abc"}}
	req := httptest.NewRequest(http.MethodGet, "/users/abc/balance", nil)
	c.Request = req
	h := &BalanceHandler{Service: &mockBalanceService{
		GetBalanceFn: func(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
			return decimal.Zero, decimal.Zero, decimal.Zero, nil
		},
	}}
	h.GetBalance(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}
}

func TestGetBalance_InvalidDate_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "user_id", Value: "1"}}
	req := httptest.NewRequest(http.MethodGet, "/users/1/balance?from=2024-01-01T00:00:00", nil) // missing Z
	c.Request = req
	h := &BalanceHandler{Service: &mockBalanceService{
		GetBalanceFn: func(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
			return decimal.Zero, decimal.Zero, decimal.Zero, nil
		},
	}}
	h.GetBalance(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}
}

func TestGetBalance_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "user_id", Value: "1"}}
	req := httptest.NewRequest(http.MethodGet, "/users/1/balance?from=2024-01-01T00:00:00Z", nil)
	c.Request = req
	h := &BalanceHandler{Service: &mockBalanceService{
		GetBalanceFn: func(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
			return decimal.Zero, decimal.Zero, decimal.Zero, shared.NewNotFound("user_transactions_not_found", "user has no transactions", nil)
		},
	}}
	h.GetBalance(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status want 404 got %d", w.Code)
	}
}

func TestGetBalance_Success_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "user_id", Value: "42"}}
	req := httptest.NewRequest(http.MethodGet, "/users/42/balance?from=2024-01-01T00:00:00Z&to=2024-12-31T23:59:59Z", nil)
	c.Request = req
	h := &BalanceHandler{Service: &mockBalanceService{
		GetBalanceFn: func(ctx context.Context, userID int64, from, to time.Time) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
			return decimal.NewFromFloat(25.21), decimal.NewFromFloat(10), decimal.NewFromFloat(15.21), nil
		},
	}}
	h.GetBalance(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d", w.Code)
	}
	var ok responses.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ok.Balance != 25.21 || ok.TotalDebits != 10 || ok.TotalCredits != 15.21 {
		t.Fatalf("payload mismatch: %+v", ok)
	}
}


