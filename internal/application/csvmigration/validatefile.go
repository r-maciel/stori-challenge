package csvmigration

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"stori-challenge/internal/domain"
	"stori-challenge/internal/ports/services"

	"github.com/shopspring/decimal"
)

// readAndValidate performs a single pass over the CSV: header check, per-row validation, and build domain transactions.
func (s *csvMigrationService) readAndValidate(r io.Reader) ([]domain.Transaction, []ParsedRow, []services.RowError, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1

	recordIdx := 0
	now := s.NowFunc()

	var (
		validTxs []domain.Transaction
		errs     []services.RowError
		rows     []ParsedRow
		seenIDs  = make(map[int64]int) // id -> firstRow
	)

	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, nil, err
		}
		recordIdx++

		// Header (mandatory): first 4 columns must match
		if recordIdx == 1 {
			if s.isHeader(rec) {
				continue
			}
			return nil, nil, nil, errors.New("invalid or missing header")
		}

		rowNum := recordIdx - 1 // data rows start at 1
		cols := len(rec)
		// Require at least the first 4 columns (ignore extras)
		if cols < len(expectedHeaders) {
			errs = append(errs, services.RowError{
				Row:     rowNum,
				Field:   "columns",
				Value:   strconv.Itoa(cols),
				Message: "at least " + strconv.Itoa(len(expectedHeaders)) + " columns required: " + strings.Join(expectedHeaders, ","),
			})
			continue
		}

		idStr := strings.TrimSpace(rec[0])
		userIDStr := strings.TrimSpace(rec[1])
		amountStr := strings.TrimSpace(rec[2])
		datetimeStr := strings.TrimSpace(rec[3])
		rows = append(rows, ParsedRow{
			RowNum:      rowNum,
			IDStr:       idStr,
			UserIDStr:   userIDStr,
			AmountStr:   amountStr,
			DatetimeStr: datetimeStr,
			Cols:        cols,
		})

		// Parse and validate
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			errs = append(errs, services.RowError{Row: rowNum, Field: "id", Value: idStr, Message: "not a valid integer"})
			continue
		}
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			errs = append(errs, services.RowError{Row: rowNum, Field: "user_id", Value: userIDStr, Message: "not a valid integer"})
			continue
		}
		amt, err := decimal.NewFromString(amountStr)
		if err != nil {
			errs = append(errs, services.RowError{Row: rowNum, Field: "amount", Value: amountStr, Message: "not a valid number"})
			continue
		}
		dt, err := time.Parse(time.RFC3339, datetimeStr)
		if err != nil {
			errs = append(errs, services.RowError{Row: rowNum, Field: "datetime", Value: datetimeStr, Message: "not a valid RFC3339 datetime"})
			continue
		}
		if dt.After(now.UTC()) {
			errs = append(errs, services.RowError{Row: rowNum, Field: "datetime", Value: datetimeStr, Message: "datetime is in the future"})
			continue
		}
		// Duplicate id within file
		if firstRow, ok := seenIDs[id]; ok {
			errs = append(errs, services.RowError{
				Row:     rowNum,
				Field:   "id",
				Value:   idStr,
				Message: "duplicate id within file (first seen at row " + strconv.Itoa(firstRow) + ")",
			})
			continue
		}
		seenIDs[id] = rowNum

		validTxs = append(validTxs, domain.Transaction{
			ID:       id,
			UserID:   userID,
			Amount:   amt,
			DateTime: dt.UTC(),
		})
	}
	return validTxs, rows, errs, nil
}

func (s *csvMigrationService) isHeader(rec []string) bool {
	if len(rec) < len(expectedHeaders) {
		return false
	}
	for i, h := range expectedHeaders {
		if strings.ToLower(strings.TrimSpace(rec[i])) != h {
			return false
		}
	}
	return true
}
