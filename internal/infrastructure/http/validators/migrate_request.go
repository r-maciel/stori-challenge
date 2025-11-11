package validators

import (
	"path/filepath"
	"stori-challenge/internal/shared"
	"strings"
)

const MaxUploadSize = 5 * 1024 * 1024 // 5MB

// ValidateFileMeta validates basic file metadata and returns an AppError
// with a specific code and message when validation fails. On success, returns nil.
func ValidateFileMeta(filename string, size int64) *shared.AppError {
	if filename == "" {
		return shared.NewBadRequest("missing_file", "file is required", nil)
	}
	if size > MaxUploadSize {
		return shared.NewBadRequest("file_too_large", "file too large", nil)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".csv" {
		return shared.NewBadRequest("wrong_extension", "file must have .csv extension", nil)
	}
	return nil
}
