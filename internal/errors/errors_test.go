package errors

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	appErr := NewAppError(http.StatusNotFound, ErrCodeNotFound, "не найден")

	if appErr.Status != http.StatusNotFound {
		t.Fatalf("Status = %d, want %d", appErr.Status, http.StatusNotFound)
	}
	if appErr.Message != "не найден" {
		t.Fatalf("Message = %q, want %q", appErr.Message, "не найден")
	}
	if appErr.Code != ErrCodeNotFound {
		t.Fatalf("Code = %v, want %v", appErr.Code, ErrCodeNotFound)
	}
}

func TestAppError_Error(t *testing.T) {
	appErr := NewAppError(http.StatusBadRequest, ErrCodeBadRequest, "плохой запрос")
	if appErr.Error() != "плохой запрос" {
		t.Fatalf("Error() = %q, want %q", appErr.Error(), "плохой запрос")
	}
}

func TestAsAppError_Success(t *testing.T) {
	original := NewAppError(http.StatusForbidden, ErrCodeForbidden, "запрещено")
	// Оборачиваем в fmt.Errorf для реалистичности
	wrapped := fmt.Errorf("service error: %w", original)

	got, ok := AsAppError(wrapped)
	if !ok {
		t.Fatal("AsAppError() returned false for wrapped AppError")
	}
	if got.Status != http.StatusForbidden {
		t.Fatalf("Status = %d, want %d", got.Status, http.StatusForbidden)
	}
}

func TestAsAppError_NotAppError(t *testing.T) {
	err := fmt.Errorf("plain error")

	_, ok := AsAppError(err)
	if ok {
		t.Fatal("AsAppError() returned true for plain error")
	}
}

func TestAsAppError_Nil(t *testing.T) {
	_, ok := AsAppError(nil)
	if ok {
		t.Fatal("AsAppError(nil) returned true")
	}
}
