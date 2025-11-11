package csvmigration

import (
	"strings"
	"testing"
	"time"
)

func Test_isHeader(t *testing.T) {
	s := &csvMigrationService{}
	ok := s.isHeader([]string{"id", "user_id", "amount", "datetime"})
	if !ok {
		t.Fatalf("expected true for exact header")
	}
	ok = s.isHeader([]string{"ID", "USER_ID", "Amount", "Datetime"})
	if !ok {
		t.Fatalf("expected true for case-insensitive header")
	}
	ok = s.isHeader([]string{"id", "amount", "user_id", "datetime"})
	if ok {
		t.Fatalf("expected false for wrong order")
	}
}

func Test_readAndValidate_AtLeast4Columns(t *testing.T) {
	now := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	s := &csvMigrationService{NowFunc: func() time.Time { return now }}
	csv := "id,user_id,amount,datetime\n1,2,3\n"
	_, _, errs, err := s.readAndValidate(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatalf("expected column count error")
	}
}
