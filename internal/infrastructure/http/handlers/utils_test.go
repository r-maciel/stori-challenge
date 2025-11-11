package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stori-challenge/internal/infrastructure/http/responses"
	"stori-challenge/internal/shared"

	"github.com/gin-gonic/gin"
)

func TestCreateErrorResponse_AppErrorKinds_MapToStatusAndBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		err        *shared.AppError
		wantStatus int
	}{
		{"bad_request", shared.NewBadRequest("invalid_input", "bad input", nil), http.StatusBadRequest},
		{"conflict", shared.NewConflict("conflict", "already exists", nil), http.StatusConflict},
		{"not_found", shared.NewNotFound("not_found", "missing", nil), http.StatusNotFound},
		{"internal", shared.NewInternal("internal", "boom", nil), http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			CreateErrorResponse(c, tc.err, map[string]string{"k": "v"})
			if w.Code != tc.wantStatus {
				t.Fatalf("status: want %d got %d", tc.wantStatus, w.Code)
			}
			var env responses.ErrorEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Error.Code != tc.err.Code || env.Error.Message != tc.err.Msg {
				t.Fatalf("body mismatch: %+v", env.Error)
			}
			if env.Error.Details == nil {
				t.Fatalf("expected details present")
			}
		})
	}
}

func TestCreateErrorResponse_NonAppError_Internal500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	CreateErrorResponse(c, assertErr("x"), []int{1})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: want 500 got %d", w.Code)
	}
	var env responses.ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code != "internal_error" || env.Error.Message != "internal error" {
		t.Fatalf("body mismatch: %+v", env.Error)
	}
	if env.Error.Details == nil {
		t.Fatalf("expected details present")
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
