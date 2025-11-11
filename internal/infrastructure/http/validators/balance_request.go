package validators

import (
	"strings"
	"time"

	"stori-challenge/internal/shared"
)

// ParseAndValidateTimeRange parses from/to in RFC3339 with 'Z' and returns a valid [from, to] window.
// Rules:
// - If both present: lower = min(from,to), upper = max(from,to)
// - If only one present: lower = provided, upper = nowUTC
// - Upper must not be greater than nowUTC
// - Inputs must end with 'Z' and parse with time.RFC3339
func ParseAndValidateTimeRange(fromStr, toStr string, nowUTC time.Time) (time.Time, time.Time, *shared.AppError) {
	nowUTC = nowUTC.UTC()
	var (
		from time.Time
		to   time.Time
	)
	hasFrom := strings.TrimSpace(fromStr) != ""
	hasTo := strings.TrimSpace(toStr) != ""

	parse := func(s string) (time.Time, *shared.AppError) {
		if !strings.HasSuffix(s, "Z") {
			return time.Time{}, shared.NewBadRequest("invalid_datetime", "datetime must be RFC3339 with Z", nil)
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, shared.NewBadRequest("invalid_datetime", "datetime must be RFC3339 with Z", err)
		}
		return t.UTC(), nil
	}

	switch {
	case hasFrom && hasTo:
		fv, err1 := parse(fromStr)
		if err1 != nil {
			return time.Time{}, time.Time{}, err1
		}
		tv, err2 := parse(toStr)
		if err2 != nil {
			return time.Time{}, time.Time{}, err2
		}
		// assign min/max
		if fv.Before(tv) || fv.Equal(tv) {
			from, to = fv, tv
		} else {
			from, to = tv, fv
		}
	case hasFrom && !hasTo:
		fv, err1 := parse(fromStr)
		if err1 != nil {
			return time.Time{}, time.Time{}, err1
		}
		from, to = fv, nowUTC
	case !hasFrom && hasTo:
		tv, err1 := parse(toStr)
		if err1 != nil {
			return time.Time{}, time.Time{}, err1
		}
		from, to = tv, nowUTC
	default:
		// No limits provided: default to full history until now.
		// Lower bound: zero time (but realistic apps might choose now - 100y). We'll use zero to include all.
		from, to = time.Time{}, nowUTC
	}

	if to.After(nowUTC) {
		return time.Time{}, time.Time{}, shared.NewBadRequest("invalid_range", "upper bound cannot be in the future", nil)
	}
	return from, to, nil
}


