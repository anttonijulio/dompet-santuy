package service

import (
	"context"
	"fmt"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) Create(ctx context.Context, userID string, req *domain.CreateCategoryRequest) (*domain.CategoryResponse, error) {
	cat := &domain.Category{
		UserID: userID,
		Name:   req.Name,
		Icon:   req.Icon,
		Color:  req.Color,
		Type:   req.Type,
	}
	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	resp := cat.ToResponse()
	return &resp, nil
}

func (s *CategoryService) ListByType(ctx context.Context, userID, typeFilter string) ([]domain.CategoryResponse, error) {
	cats, err := s.categoryRepo.FindByUserID(ctx, userID, typeFilter)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	result := make([]domain.CategoryResponse, len(cats))
	for i, c := range cats {
		result[i] = c.ToResponse()
	}
	return result, nil
}
