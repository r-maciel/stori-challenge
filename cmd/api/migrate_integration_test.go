package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	infradb "stori-challenge/internal/infrastructure/db"
	"stori-challenge/internal/infrastructure/http/responses"
	testinfra "stori-challenge/internal/shared/test"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := testinfra.OpenTestDB(t.Name())
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Use real cmd/api wiring to exercise this package
	r := NewServerWithDB(db)
	return r, db
}

func makeMultipartCSV(t *testing.T, filename string, csv string) (contentType string, body *bytes.Buffer) {
	t.Helper()
	body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewBufferString(csv)); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return writer.FormDataContentType(), body
}

func TestMigrateIntegration_Success201(t *testing.T) {
	router, db := newTestRouter(t)
	// two valid rows
	csv := "id,user_id,amount,datetime\n" +
		"10001,1,1.23,2023-01-01T00:00:00Z\n" +
		"10002,2,4.56,2023-01-02T00:00:00Z\n"
	ct, body := makeMultipartCSV(t, "data.csv", csv)

	req := httptest.NewRequest(http.MethodPost, "/v1/migrate", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status: want 201 got %d; body=%s", w.Code, w.Body.String())
	}
	var ok responses.MigrateSuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ok.Inserted != 2 {
		t.Fatalf("inserted want 2 got %d", ok.Inserted)
	}
	// Verify in DB within the same test transaction
	repo := infradb.NewTransactionRepo(db)
	exists, err := repo.ExistsByIDs(context.Background(), []int64{10001, 10002})
	if err != nil {
		t.Fatalf("exists check: %v", err)
	}
	if len(exists) != 2 || !exists[10001] || !exists[10002] {
		t.Fatalf("rows not found in db: %v", exists)
	}
}

func TestMigrateIntegration_Conflict409(t *testing.T) {
	router, db := newTestRouter(t)
	// first request inserts id=20001
	csv1 := "id,user_id,amount,datetime\n" +
		"20001,1,1.00,2023-01-01T00:00:00Z\n"
	ct1, body1 := makeMultipartCSV(t, "data.csv", csv1)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/migrate", body1)
	req1.Header.Set("Content-Type", ct1)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("pre-insert status want 201 got %d; body=%s", w1.Code, w1.Body.String())
	}

	// second request tries to insert same id → conflict
	ct2, body2 := makeMultipartCSV(t, "data.csv", csv1)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/migrate", body2)
	req2.Header.Set("Content-Type", ct2)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("status: want 409 got %d; body=%s", w2.Code, w2.Body.String())
	}
	var env responses.ErrorEnvelope
	if err := json.Unmarshal(w2.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code == "" {
		t.Fatalf("expected error code set")
	}
	// Verify only the first row exists
	repo := infradb.NewTransactionRepo(db)
	exists, err := repo.ExistsByIDs(context.Background(), []int64{20001})
	if err != nil {
		t.Fatalf("exists check: %v", err)
	}
	if len(exists) != 1 || !exists[20001] {
		t.Fatalf("expected id 20001 to exist, got %v", exists)
	}
}

func TestMigrateIntegration_Validation400(t *testing.T) {
	router, db := newTestRouter(t)
	// invalid: date in the future
	future := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	csv := "id,user_id,amount,datetime\n" +
		"30001,1,1.00," + future + "\n"
	ct, body := makeMultipartCSV(t, "data.csv", csv)
	req := httptest.NewRequest(http.MethodPost, "/v1/migrate", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: want 400 got %d; body=%s", w.Code, w.Body.String())
	}
	var env responses.ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %s", env.Error.Code)
	}
	// Verify no insert was made
	repo := infradb.NewTransactionRepo(db)
	exists, err := repo.ExistsByIDs(context.Background(), []int64{30001})
	if err != nil {
		t.Fatalf("exists check: %v", err)
	}
	if len(exists) != 0 {
		t.Fatalf("expected no rows, got %v", exists)
	}
}
