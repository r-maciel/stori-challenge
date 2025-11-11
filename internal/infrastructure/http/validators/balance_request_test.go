package validators

import (
	"testing"
	"time"
)

func TestParseAndValidateTimeRange_BothProvided_OrdersBounds(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	fromStr := "2024-01-01T00:00:00Z"
	toStr := "2024-12-31T23:59:59Z"
	from, to, err := ParseAndValidateTimeRange(toStr, fromStr, now) // swapped on purpose
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !from.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("from mismatch: %v", from)
	}
	if !to.Equal(time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)) {
		t.Fatalf("to mismatch: %v", to)
	}
}

func TestParseAndValidateTimeRange_OnlyFrom_UpperIsNow(t *testing.T) {
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	fromStr := "2025-05-01T00:00:00Z"
	gotFrom, gotTo, err := ParseAndValidateTimeRange(fromStr, "", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotFrom.Equal(time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("from mismatch: %v", gotFrom)
	}
	if !gotTo.Equal(now) {
		t.Fatalf("to should be now: %v", gotTo)
	}
}

func TestParseAndValidateTimeRange_OnlyTo_TreatedAsLowerBound(t *testing.T) {
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	toStr := "2025-05-01T00:00:00Z"
	gotFrom, gotTo, err := ParseAndValidateTimeRange("", toStr, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotFrom.Equal(time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("from mismatch: %v", gotFrom)
	}
	if !gotTo.Equal(now) {
		t.Fatalf("to should be now: %v", gotTo)
	}
}

func TestParseAndValidateTimeRange_InvalidFormat_NoZ(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	_, _, err := ParseAndValidateTimeRange("2024-01-01T00:00:00", "", now)
	if err == nil {
		t.Fatalf("expected error for missing Z")
	}
}

func TestParseAndValidateTimeRange_UpperAfterNow_IsBadRequest(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	from := "2024-12-30T00:00:00Z"
	to := "2025-12-31T00:00:00Z" // future relative to now
	_, _, err := ParseAndValidateTimeRange(from, to, now)
	if err == nil {
		t.Fatalf("expected error for future upper bound")
	}
}


