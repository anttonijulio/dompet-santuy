package handler

import (
	"errors"
	"strings"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/middleware"
	"github.com/antonidev/dompet-santuy/internal/repository"
	"github.com/antonidev/dompet-santuy/internal/response"
	"github.com/antonidev/dompet-santuy/internal/service"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req domain.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.authService.Register(c.Request().Context(), &req)
	if errors.Is(err, repository.ErrDuplicateEmail) {
		return response.Conflict(c, "email already registered")
	}
	if err != nil {
		return response.InternalServerError(c, "registration failed")
	}

	return response.Created(c, "registration successful", user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	tokens, err := h.authService.Login(c.Request().Context(), &req)
	if errors.Is(err, service.ErrInvalidCredentials) {
		return response.Unauthorized(c, "invalid email or password")
	}
	if err != nil {
		return response.InternalServerError(c, "login failed")
	}

	return response.OK(c, "login successful", tokens)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req domain.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	tokens, err := h.authService.Refresh(c.Request().Context(), req.RefreshToken)
	if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrTokenRevoked) {
		return response.Unauthorized(c, err.Error())
	}
	if err != nil {
		return response.InternalServerError(c, "token refresh failed")
	}

	return response.OK(c, "token refreshed", tokens)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req domain.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	_ = h.authService.Logout(c.Request().Context(), req.RefreshToken)
	return response.NoContent(c, "logged out successfully")
}

func (h *AuthHandler) LogoutAll(c echo.Context) error {
	userID := c.Get(middleware.UserIDKey).(string)

	_ = h.authService.LogoutAll(c.Request().Context(), userID)
	return response.NoContent(c, "all sessions terminated")
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID := c.Get(middleware.UserIDKey).(string)

	user, err := h.authService.GetProfile(c.Request().Context(), userID)
	if errors.Is(err, repository.ErrNotFound) {
		return response.NotFound(c, "user not found")
	}
	if err != nil {
		return response.InternalServerError(c, "failed to get profile")
	}

	return response.OK(c, "profile retrieved", user)
}
