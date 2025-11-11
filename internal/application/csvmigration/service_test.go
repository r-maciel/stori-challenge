package csvmigration

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"stori-challenge/internal/domain"
	"stori-challenge/internal/ports/repositories"
	"stori-challenge/internal/shared"
)

type fakeRepo struct {
	exists    map[int64]bool
	existsErr error
	captured  []domain.Transaction
	bulkErr   error
}

func (f *fakeRepo) ExistsByIDs(ctx context.Context, ids []int64) (map[int64]bool, error) {
	if f.existsErr != nil {
		return nil, f.existsErr
	}
	out := make(map[int64]bool, len(f.exists))
	for k, v := range f.exists {
		out[k] = v
	}
	return out, nil
}

func (f *fakeRepo) BulkInsert(ctx context.Context, txs []domain.Transaction) error {
	f.captured = append(f.captured, txs...)
	return f.bulkErr
}

func newSvcWithRepo(t *testing.T, repo repositories.TransactionRepository, now time.Time) *csvMigrationService {
	t.Helper()
	svc := NewCsvMigrationService(repo).(*csvMigrationService)
	svc.NowFunc = func() time.Time { return now }
	return svc
}

func r(s string) io.Reader { return strings.NewReader(s) }

func TestProcess_HappyPath(t *testing.T) {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	svc := newSvcWithRepo(t, repo, now)

	csv := "id,user_id,amount,datetime\n1,10,12.34,2024-06-01T00:00:00Z\n2,20,-5.00,2024-06-02T00:00:00Z\n"
	inserted, items, err := svc.Process(context.Background(), r(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no items, got %v", items)
	}
	if inserted != 2 {
		t.Fatalf("expected inserted 2, got %d", inserted)
	}
	if len(repo.captured) != 2 {
		t.Fatalf("expected repo captured 2, got %d", len(repo.captured))
	}
}

func TestProcess_HeaderInvalid(t *testing.T) {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	svc := newSvcWithRepo(t, repo, now)

	csv := "bad,header,here,now\n"
	inserted, items, err := svc.Process(context.Background(), r(csv))
	if inserted != 0 {
		t.Fatalf("expected inserted 0, got %d", inserted)
	}
	var ae *shared.AppError
	if !errors.As(err, &ae) || ae.Kind != shared.BadRequestKind {
		t.Fatalf("expected BadRequest AppError, got %v", err)
	}
	if len(items) != 1 || items[0].Field != "file" {
		t.Fatalf("expected one file-level error item, got %v", items)
	}
}

func TestProcess_ConflictIDs(t *testing.T) {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{exists: map[int64]bool{1: true}}
	svc := newSvcWithRepo(t, repo, now)

	csv := "id,user_id,amount,datetime\n1,10,12.34,2024-06-01T00:00:00Z\n"
	inserted, items, err := svc.Process(context.Background(), r(csv))
	if inserted != 0 {
		t.Fatalf("expected inserted 0, got %d", inserted)
	}
	var ae *shared.AppError
	if !errors.As(err, &ae) || ae.Kind != shared.ConflictKind {
		t.Fatalf("expected Conflict AppError, got %v", err)
	}
	if len(items) != 1 || items[0].Field != "id" {
		t.Fatalf("expected one id conflict item, got %v", items)
	}
}

func TestProcess_DBErrorOnExists(t *testing.T) {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{existsErr: errors.New("db down")}
	svc := newSvcWithRepo(t, repo, now)
	csv := "id,user_id,amount,datetime\n1,10,12.34,2024-06-01T00:00:00Z\n"
	_, _, err := svc.Process(context.Background(), r(csv))
	var ae *shared.AppError
	if !errors.As(err, &ae) || ae.Kind != shared.InternalKind {
		t.Fatalf("expected Internal AppError, got %v", err)
	}
}

func TestProcess_ValidationErrors(t *testing.T) {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	svc := newSvcWithRepo(t, repo, now)

	// amount invalid, datetime future
	csv := "id,user_id,amount,datetime\n1,10,xx,2025-06-01T00:00:00Z\n"
	_, items, err := svc.Process(context.Background(), r(csv))
	var ae *shared.AppError
	if !errors.As(err, &ae) || ae.Kind != shared.BadRequestKind {
		t.Fatalf("expected BadRequest AppError, got %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected validation items, got none")
	}
}
