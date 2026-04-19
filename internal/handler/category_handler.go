package handler

import (
	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/middleware"
	"github.com/antonidev/dompet-santuy/internal/response"
	"github.com/antonidev/dompet-santuy/internal/service"
	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	categorySvc *service.CategoryService
}

func NewCategoryHandler(categorySvc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categorySvc: categorySvc}
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req domain.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userID := c.Get(middleware.UserIDKey).(string)
	cat, err := h.categorySvc.Create(c.Request().Context(), userID, &req)
	if err != nil {
		return response.InternalServerError(c, "failed to create category")
	}

	return response.Created(c, "category created", cat)
}

func (h *CategoryHandler) List(c echo.Context) error {
	userID := c.Get(middleware.UserIDKey).(string)
	typeFilter := c.QueryParam("type")

	cats, err := h.categorySvc.ListByType(c.Request().Context(), userID, typeFilter)
	if err != nil {
		return response.InternalServerError(c, "failed to list categories")
	}

	return response.OK(c, "categories retrieved", cats)
}
