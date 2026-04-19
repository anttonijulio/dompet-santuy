package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/google/uuid"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	tx.ID = uuid.NewString()
	query := `INSERT INTO transactions (id, user_id, category_id, amount, type, note, date) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, tx.ID, tx.UserID, tx.CategoryID, tx.Amount, tx.Type, nullableString(tx.Note), tx.Date)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) FindByUserID(ctx context.Context, userID string, f domain.ListTransactionsFilter) ([]domain.Transaction, int, error) {
	where := `t.user_id = ?`
	args := []any{userID}

	if f.StartDate != "" {
		where += ` AND t.date >= ?`
		args = append(args, f.StartDate)
	}
	if f.EndDate != "" {
		where += ` AND t.date <= ?`
		args = append(args, f.EndDate)
	}
	if f.Type != "" {
		where += ` AND t.type = ?`
		args = append(args, f.Type)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM transactions t WHERE ` + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	listQuery := `
		SELECT t.id, t.user_id, t.category_id, t.amount, t.type, COALESCE(t.note,''), t.date, t.created_at,
		       c.id, c.name, COALESCE(c.icon,''), COALESCE(c.color,''), c.type
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE ` + where + `
		ORDER BY t.date DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, listQuery, append(args, limit, f.Offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var cat domain.Category
		err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.CategoryID, &tx.Amount, &tx.Type, &tx.Note, &tx.Date, &tx.CreatedAt,
			&cat.ID, &cat.Name, &cat.Icon, &cat.Color, &cat.Type,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		tx.Category = &cat
		txs = append(txs, tx)
	}
	return txs, total, rows.Err()
}

func (r *TransactionRepository) FindByIDAndUserID(ctx context.Context, id, userID string) (*domain.Transaction, error) {
	query := `
		SELECT t.id, t.user_id, t.category_id, t.amount, t.type, COALESCE(t.note,''), t.date, t.created_at,
		       c.id, c.name, COALESCE(c.icon,''), COALESCE(c.color,''), c.type
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.id = ? AND t.user_id = ?`

	var tx domain.Transaction
	var cat domain.Category
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&tx.ID, &tx.UserID, &tx.CategoryID, &tx.Amount, &tx.Type, &tx.Note, &tx.Date, &tx.CreatedAt,
		&cat.ID, &cat.Name, &cat.Icon, &cat.Color, &cat.Type,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find transaction: %w", err)
	}
	tx.Category = &cat
	return &tx, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	query := `UPDATE transactions SET category_id = ?, amount = ?, type = ?, note = ?, date = ? WHERE id = ? AND user_id = ?`
	res, err := r.db.ExecContext(ctx, query, tx.CategoryID, tx.Amount, tx.Type, nullableString(tx.Note), tx.Date, tx.ID, tx.UserID)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update transaction rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM transactions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete transaction rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
