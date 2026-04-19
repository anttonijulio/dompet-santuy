package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/google/uuid"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *domain.Category) error {
	cat.ID = uuid.NewString()
	query := `INSERT INTO categories (id, user_id, name, icon, color, type) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, cat.ID, cat.UserID, cat.Name, cat.Icon, cat.Color, cat.Type)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) FindByUserID(ctx context.Context, userID, typeFilter string) ([]domain.Category, error) {
	query := `SELECT id, user_id, name, COALESCE(icon,''), COALESCE(color,''), type, created_at
	          FROM categories WHERE user_id = ?`
	args := []any{userID}

	if typeFilter != "" {
		query += ` AND type = ?`
		args = append(args, typeFilter)
	}
	query += ` ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Icon, &c.Color, &c.Type, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) FindByIDAndUserID(ctx context.Context, id, userID string) (*domain.Category, error) {
	query := `SELECT id, user_id, name, COALESCE(icon,''), COALESCE(color,''), type, created_at
	          FROM categories WHERE id = ? AND user_id = ?`
	row := r.db.QueryRowContext(ctx, query, id, userID)

	var c domain.Category
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Icon, &c.Color, &c.Type, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find category: %w", err)
	}
	return &c, nil
}
