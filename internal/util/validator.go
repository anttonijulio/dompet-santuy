package util

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/labstack/echo/v4"
)

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
