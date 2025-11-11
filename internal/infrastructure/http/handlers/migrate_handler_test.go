package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stori-challenge/internal/infrastructure/http/responses"
	"stori-challenge/internal/ports/services"
	"stori-challenge/internal/shared"

	"github.com/gin-gonic/gin"
)

type mockMigrationService struct {
	ProcessFn func(ctx context.Context, r io.Reader) (int, []services.RowError, error)
}

func (m *mockMigrationService) Process(ctx context.Context, r io.Reader) (int, []services.RowError, error) {
	return m.ProcessFn(ctx, r)
}

func TestPostMigrate_MissingFile_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/migrate", nil)
	h := &MigrateHandler{Service: &mockMigrationService{
		ProcessFn: func(ctx context.Context, r io.Reader) (int, []services.RowError, error) { return 0, nil, nil },
	}}
	h.PostMigrate(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}
}

func TestPostMigrate_WrongExtension_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "data.txt")
	_, _ = io.Copy(part, strings.NewReader("id,user_id,amount,datetime\n"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/migrate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h := &MigrateHandler{Service: &mockMigrationService{
		ProcessFn: func(ctx context.Context, r io.Reader) (int, []services.RowError, error) { return 0, nil, nil },
	}}
	h.PostMigrate(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}
}

func TestPostMigrate_ServiceError_MapsRowErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "data.csv")
	_, _ = io.Copy(part, strings.NewReader("id,user_id,amount,datetime\n"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/migrate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	rowErrs := []services.RowError{
		{Row: 2, Field: "amount", Value: "abc", Message: "must be a number"},
		{Row: 3, Field: "id", Value: "1", Message: "duplicate"},
	}
	h := &MigrateHandler{Service: &mockMigrationService{
		ProcessFn: func(ctx context.Context, r io.Reader) (int, []services.RowError, error) {
			return 0, rowErrs, shared.NewConflict("conflict", "conflict", nil)
		},
	}}
	h.PostMigrate(c)
	if w.Code != http.StatusConflict {
		t.Fatalf("status want 409 got %d", w.Code)
	}
	var env responses.ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// details should be an array of MigrateRowError; we just check it's present
	if env.Error.Details == nil {
		t.Fatalf("expected error details")
	}
}

func TestPostMigrate_Success_Returns201(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "data.csv")
	_, _ = io.Copy(part, strings.NewReader("id,user_id,amount,datetime\n"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/migrate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h := &MigrateHandler{Service: &mockMigrationService{
		ProcessFn: func(ctx context.Context, r io.Reader) (int, []services.RowError, error) {
			return 42, nil, nil
		},
	}}
	h.PostMigrate(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status want 201 got %d", w.Code)
	}
	var ok responses.MigrateSuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ok); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ok.Inserted != 42 {
		t.Fatalf("inserted want 42 got %d", ok.Inserted)
	}
}
