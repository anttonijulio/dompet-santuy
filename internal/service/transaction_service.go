package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/repository"
)

var ErrCategoryNotOwned = errors.New("category not found or does not belong to user")
var ErrCategoryTypeMismatch = errors.New("category type does not match transaction type")

var dateFormats = []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"}

type TransactionService struct {
	transactionRepo *repository.TransactionRepository
	categoryRepo    *repository.CategoryRepository
}

func NewTransactionService(transactionRepo *repository.TransactionRepository, categoryRepo *repository.CategoryRepository) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

func (s *TransactionService) Create(ctx context.Context, userID string, req *domain.CreateTransactionRequest) (*domain.TransactionResponse, error) {
	cat, err := s.categoryRepo.FindByIDAndUserID(ctx, req.CategoryID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCategoryNotOwned
	}
	if err != nil {
		return nil, fmt.Errorf("find category: %w", err)
	}

	if cat.Type != req.Type {
		return nil, ErrCategoryTypeMismatch
	}

	date, err := parseDate(req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	tx := &domain.Transaction{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Type:       req.Type,
		Note:       req.Note,
		Date:       date,
		Category:   cat,
	}
	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	resp := tx.ToResponse()
	return &resp, nil
}

func (s *TransactionService) GetByID(ctx context.Context, userID, transactionID string) (*domain.TransactionResponse, error) {
	tx, err := s.transactionRepo.FindByIDAndUserID(ctx, transactionID, userID)
	if err != nil {
		return nil, err
	}
	resp := tx.ToResponse()
	return &resp, nil
}

func (s *TransactionService) Update(ctx context.Context, userID, transactionID string, req *domain.UpdateTransactionRequest) (*domain.TransactionResponse, error) {
	tx, err := s.transactionRepo.FindByIDAndUserID(ctx, transactionID, userID)
	if err != nil {
		return nil, err
	}

	cat, err := s.categoryRepo.FindByIDAndUserID(ctx, req.CategoryID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCategoryNotOwned
	}
	if err != nil {
		return nil, fmt.Errorf("find category: %w", err)
	}
	if cat.Type != req.Type {
		return nil, ErrCategoryTypeMismatch
	}

	date, err := parseDate(req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	tx.CategoryID = req.CategoryID
	tx.Amount = req.Amount
	tx.Type = req.Type
	tx.Note = req.Note
	tx.Date = date
	tx.Category = cat

	if err := s.transactionRepo.Update(ctx, tx); err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	resp := tx.ToResponse()
	return &resp, nil
}

func (s *TransactionService) Delete(ctx context.Context, userID, transactionID string) error {
	if _, err := s.transactionRepo.FindByIDAndUserID(ctx, transactionID, userID); err != nil {
		return err
	}
	return s.transactionRepo.Delete(ctx, transactionID, userID)
}

func (s *TransactionService) List(ctx context.Context, userID string, f domain.ListTransactionsFilter) ([]domain.TransactionResponse, int, error) {
	txs, total, err := s.transactionRepo.FindByUserID(ctx, userID, f)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}

	result := make([]domain.TransactionResponse, len(txs))
	for i, tx := range txs {
		result[i] = tx.ToResponse()
	}
	return result, total, nil
}

func parseDate(s string) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as date", s)
}
