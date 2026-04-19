package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Response[T any] struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    T        `json:"data,omitempty"`
	Errors  []string `json:"errors,omitempty"`
	Meta    *Meta    `json:"meta,omitempty"`
}

type ListResponse[T any] struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    []T      `json:"data"`
	Errors  []string `json:"errors,omitempty"`
	Meta    *Meta    `json:"meta,omitempty"`
}

func OK[T any](c echo.Context, message string, data T) error {
	return c.JSON(http.StatusOK, Response[T]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created[T any](c echo.Context, message string, data T) error {
	return c.JSON(http.StatusCreated, Response[T]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func List[T any](c echo.Context, message string, data []T) error {
	if data == nil {
		data = []T{}
	}
	return c.JSON(http.StatusOK, ListResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Paginated[T any](c echo.Context, message string, data []T, meta Meta) error {
	if data == nil {
		data = []T{}
	}
	return c.JSON(http.StatusOK, ListResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, Response[any]{Success: false, Message: message})
}

func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, Response[any]{Success: false, Message: message})
}

func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, Response[any]{Success: false, Message: message})
}

func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, Response[any]{Success: false, Message: message})
}

func Conflict(c echo.Context, message string) error {
	return c.JSON(http.StatusConflict, Response[any]{Success: false, Message: message})
}

func UnprocessableEntity(c echo.Context, message string) error {
	return c.JSON(http.StatusUnprocessableEntity, Response[any]{Success: false, Message: message})
}

func InternalServerError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, Response[any]{Success: false, Message: message})
}

func NoContent(c echo.Context, message string) error {
	return c.JSON(http.StatusOK, Response[any]{
		Success: true,
		Message: message,
	})
}

func NewMeta(page, limit, total int) Meta {
	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}
	return Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
