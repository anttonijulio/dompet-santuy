package util

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/labstack/echo/v4"
)

var txDateFormats = []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"}

func validateTransactionDate(s string) (ok bool, errMsg string) {
	if strings.TrimSpace(s) == "" {
		return false, "date is required"
	}
	var t time.Time
	for _, f := range txDateFormats {
		if parsed, err := time.Parse(f, s); err == nil {
			t = parsed.UTC()
			break
		}
	}
	if t.IsZero() {
		return false, "date is invalid (expected YYYY-MM-DD)"
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if t.Truncate(24 * time.Hour).After(today) {
		return false, "date must not be in the future"
	}
	return true, ""
}

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Validator struct{}

func NewValidator() *Validator { return &Validator{} }

func (v *Validator) Validate(i interface{}) error {
	var errs []string

	switch req := i.(type) {
	case *domain.RegisterRequest:
		if strings.TrimSpace(req.Name) == "" {
			errs = append(errs, "name is required")
		} else if len(req.Name) < 2 || len(req.Name) > 100 {
			errs = append(errs, "name must be between 2 and 100 characters")
		}
		if !emailRe.MatchString(req.Email) {
			errs = append(errs, "email is invalid")
		}
		if len(req.Password) < 8 {
			errs = append(errs, "password must be at least 8 characters")
		}

	case *domain.LoginRequest:
		if !emailRe.MatchString(req.Email) {
			errs = append(errs, "email is invalid")
		}
		if req.Password == "" {
			errs = append(errs, "password is required")
		}

	case *domain.RefreshRequest:
		if strings.TrimSpace(req.RefreshToken) == "" {
			errs = append(errs, "refresh_token is required")
		}

	case *domain.CreateCategoryRequest:
		if strings.TrimSpace(req.Name) == "" {
			errs = append(errs, "name is required")
		} else if len(req.Name) > 100 {
			errs = append(errs, "name must be at most 100 characters")
		}
		if req.Type != "income" && req.Type != "expense" {
			errs = append(errs, "type must be income or expense")
		}

	case *domain.UpdateCategoryRequest:
		if strings.TrimSpace(req.Name) == "" {
			errs = append(errs, "name is required")
		} else if len(req.Name) > 100 {
			errs = append(errs, "name must be at most 100 characters")
		}
		if req.Type != "income" && req.Type != "expense" {
			errs = append(errs, "type must be income or expense")
		}

	case *domain.CreateTransactionRequest:
		if strings.TrimSpace(req.CategoryID) == "" {
			errs = append(errs, "category_id is required")
		}
		if req.Amount <= 0 {
			errs = append(errs, "amount must be greater than 0")
		}
		if req.Type != "income" && req.Type != "expense" {
			errs = append(errs, "type must be income or expense")
		}
		if ok, msg := validateTransactionDate(req.Date); !ok {
			errs = append(errs, msg)
		}

	case *domain.UpdateTransactionRequest:
		if strings.TrimSpace(req.CategoryID) == "" {
			errs = append(errs, "category_id is required")
		}
		if req.Amount <= 0 {
			errs = append(errs, "amount must be greater than 0")
		}
		if req.Type != "income" && req.Type != "expense" {
			errs = append(errs, "type must be income or expense")
		}
		if ok, msg := validateTransactionDate(req.Date); !ok {
			errs = append(errs, msg)
		}
	}

	if len(errs) > 0 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": "validation failed",
			"errors":  errs,
		})
	}
	return nil
}

// FormatValidationError converts validation errors to a consistent API response.
func FormatValidationError(err error) string {
	return fmt.Sprintf("validation error: %v", err)
}
