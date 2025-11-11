package csvmigration

import (
	"context"
	"testing"
	"time"

	"stori-challenge/internal/domain"
)

func Test_insertAll_CallsRepo(t *testing.T) {
	repo := &fakeRepo{}
	s := &csvMigrationService{Repo: repo, NowFunc: func() time.Time { return time.Now().UTC() }}
	txs := []domain.Transaction{{ID: 1}, {ID: 2}}
	if err := s.insertAll(context.Background(), txs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.captured) != 2 {
		t.Fatalf("expected 2 captured, got %d", len(repo.captured))
	}
}
