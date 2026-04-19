package handler

import (
	"errors"
	"strconv"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/middleware"
	"github.com/antonidev/dompet-santuy/internal/response"
	"github.com/antonidev/dompet-santuy/internal/service"
	"github.com/labstack/echo/v4"
)

type TransactionHandler struct {
	transactionSvc *service.TransactionService
}

func NewTransactionHandler(transactionSvc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionSvc: transactionSvc}
}

func (h *TransactionHandler) Create(c echo.Context) error {
	var req domain.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userID := c.Get(middleware.UserIDKey).(string)
	tx, err := h.transactionSvc.Create(c.Request().Context(), userID, &req)
	if errors.Is(err, service.ErrCategoryNotOwned) {
		return response.UnprocessableEntity(c, "category not found or does not belong to user")
	}
	if errors.Is(err, service.ErrCategoryTypeMismatch) {
		return response.UnprocessableEntity(c, "category type does not match transaction type")
	}
	if err != nil {
		return response.InternalServerError(c, "failed to create transaction")
	}

	return response.Created(c, "transaction created", tx)
}

func (h *TransactionHandler) List(c echo.Context) error {
	userID := c.Get(middleware.UserIDKey).(string)

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	filter := domain.ListTransactionsFilter{
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
		Type:      c.QueryParam("type"),
		Limit:     limit,
		Offset:    offset,
	}

	txs, total, err := h.transactionSvc.List(c.Request().Context(), userID, filter)
	if err != nil {
		return response.InternalServerError(c, "failed to list transactions")
	}

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	page := (filter.Offset / filter.Limit) + 1
	meta := response.NewMeta(page, filter.Limit, total)

	return response.Paginated(c, "transactions retrieved", txs, meta)
}
