package csvmigration

import (
	"context"
	"testing"
	"time"

	"stori-challenge/internal/domain"
)

func Test_checkConflicts_ReturnsMap(t *testing.T) {
	repo := &fakeRepo{exists: map[int64]bool{1: true, 3: true}}
	s := &csvMigrationService{Repo: repo, NowFunc: func() time.Time { return time.Now().UTC() }}
	txs := []domain.Transaction{{ID: 1}, {ID: 2}, {ID: 3}}
	m, err := s.checkConflicts(context.Background(), txs)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !m[1] || !m[3] || m[2] {
		t.Fatalf("unexpected map: %v", m)
	}
}

func Test_buildConflictErrors_MapsRows(t *testing.T) {
	s := &csvMigrationService{}
	rows := []ParsedRow{
		{RowNum: 1, IDStr: "1"},
		{RowNum: 2, IDStr: "2"},
	}
	existing := map[int64]bool{2: true}
	errs := s.buildConflictErrors(rows, existing)
	if len(errs) != 1 || errs[0].Row != 2 || errs[0].Field != "id" {
		t.Fatalf("unexpected errs: %v", errs)
	}
}
