package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/google/uuid"
)

var ErrCategoryInUse = errors.New("category has linked transactions")

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

func (r *CategoryRepository) Update(ctx context.Context, cat *domain.Category) error {
	query := `UPDATE categories SET name = ?, icon = ?, color = ?, type = ? WHERE id = ? AND user_id = ?`
	res, err := r.db.ExecContext(ctx, query, cat.Name, cat.Icon, cat.Color, cat.Type, cat.ID, cat.UserID)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update category rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		if isFKViolation(err) {
			return ErrCategoryInUse
		}
		return fmt.Errorf("delete category: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete category rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func isFKViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "1451") || strings.Contains(err.Error(), "foreign key constraint"))
}
