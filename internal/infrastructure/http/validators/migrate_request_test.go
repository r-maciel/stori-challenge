package validators

import (
	"testing"

	"stori-challenge/internal/shared"
)

func TestValidateFileMeta_EmptyFilename_ReturnsMissingFile(t *testing.T) {
	err := ValidateFileMeta("", 0)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Kind != shared.BadRequestKind || err.Code != "missing_file" {
		t.Fatalf("unexpected app error: kind=%s code=%s", err.Kind, err.Code)
	}
}

func TestValidateFileMeta_SizeTooLarge_ReturnsFileTooLarge(t *testing.T) {
	err := ValidateFileMeta("data.csv", MaxUploadSize+1)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Kind != shared.BadRequestKind || err.Code != "file_too_large" {
		t.Fatalf("unexpected app error: kind=%s code=%s", err.Kind, err.Code)
	}
}

func TestValidateFileMeta_WrongExtension_ReturnsWrongExtension(t *testing.T) {
	err := ValidateFileMeta("data.txt", 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Kind != shared.BadRequestKind || err.Code != "wrong_extension" {
		t.Fatalf("unexpected app error: kind=%s code=%s", err.Kind, err.Code)
	}
}

func TestValidateFileMeta_UppercaseCSV_Accepted(t *testing.T) {
	err := ValidateFileMeta("data.CSV", 10)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateFileMeta_ExactlyMaxSize_Accepted(t *testing.T) {
	err := ValidateFileMeta("data.csv", MaxUploadSize)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
