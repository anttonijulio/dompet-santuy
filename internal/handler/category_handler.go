package handler

import (
	"errors"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/middleware"
	"github.com/antonidev/dompet-santuy/internal/repository"
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

	return response.List(c, "categories retrieved", cats)
}

func (h *CategoryHandler) Get(c echo.Context) error {
	categoryID := c.Param("id")
	userID := c.Get(middleware.UserIDKey).(string)

	cat, err := h.categorySvc.GetByID(c.Request().Context(), userID, categoryID)
	if errors.Is(err, repository.ErrNotFound) {
		return response.NotFound(c, "category not found")
	}
	if err != nil {
		return response.InternalServerError(c, "failed to get category")
	}

	return response.OK(c, "category retrieved", cat)
}

func (h *CategoryHandler) Update(c echo.Context) error {
	categoryID := c.Param("id")
	var req domain.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userID := c.Get(middleware.UserIDKey).(string)
	cat, err := h.categorySvc.Update(c.Request().Context(), userID, categoryID, &req)
	if errors.Is(err, repository.ErrNotFound) {
		return response.NotFound(c, "category not found")
	}
	if err != nil {
		return response.InternalServerError(c, "failed to update category")
	}

	return response.OK(c, "category updated", cat)
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	categoryID := c.Param("id")
	userID := c.Get(middleware.UserIDKey).(string)

	err := h.categorySvc.Delete(c.Request().Context(), userID, categoryID)
	if errors.Is(err, repository.ErrNotFound) {
		return response.NotFound(c, "category not found")
	}
	if errors.Is(err, repository.ErrCategoryInUse) {
		return response.Conflict(c, "category has linked transactions and cannot be deleted")
	}
	if err != nil {
		return response.InternalServerError(c, "failed to delete category")
	}

	return response.NoContent(c, "category deleted")
}
