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
	if f.CategoryID != "" {
		where += ` AND t.category_id = ?`
		args = append(args, f.CategoryID)
	}
	if f.CategoryType != "" {
		where += ` AND c.type = ?`
		args = append(args, f.CategoryType)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM transactions t JOIN categories c ON t.category_id = c.id WHERE ` + where
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

func (r *TransactionRepository) GetSummary(ctx context.Context, userID string, f domain.SummaryFilter) (*domain.TransactionSummary, error) {
	where := `user_id = ?`
	args := []any{userID}
	if f.StartDate != "" {
		where += ` AND date >= ?`
		args = append(args, f.StartDate)
	}
	if f.EndDate != "" {
		where += ` AND date <= ?`
		args = append(args, f.EndDate)
	}

	var s domain.TransactionSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN type='income'  THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0),
			COUNT(CASE WHEN type='income'  THEN 1 END),
			COUNT(CASE WHEN type='expense' THEN 1 END),
			COALESCE(CAST(AVG(CASE WHEN type='income'  THEN amount END) AS SIGNED), 0),
			COALESCE(CAST(AVG(CASE WHEN type='expense' THEN amount END) AS SIGNED), 0)
		FROM transactions WHERE `+where, args...).Scan(
		&s.TotalCount, &s.TotalIncome, &s.TotalExpense,
		&s.IncomeCount, &s.ExpenseCount,
		&s.AvgIncome, &s.AvgExpense,
	)
	if err != nil {
		return nil, fmt.Errorf("summary aggregates: %w", err)
	}
	s.Balance = s.TotalIncome - s.TotalExpense

	for _, txType := range []string{"income", "expense"} {
		cats, err := r.topCategories(ctx, userID, txType, f)
		if err != nil {
			return nil, err
		}
		if txType == "income" {
			s.TopIncomeCategories = cats
		} else {
			s.TopExpenseCategories = cats
		}
	}
	return &s, nil
}

func (r *TransactionRepository) topCategories(ctx context.Context, userID, txType string, f domain.SummaryFilter) ([]domain.CategorySummary, error) {
	where := `t.user_id = ? AND t.type = ?`
	args := []any{userID, txType}
	if f.StartDate != "" {
		where += ` AND t.date >= ?`
		args = append(args, f.StartDate)
	}
	if f.EndDate != "" {
		where += ` AND t.date <= ?`
		args = append(args, f.EndDate)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.name, COALESCE(c.icon,''), COALESCE(c.color,''), COUNT(*), SUM(t.amount)
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE `+where+`
		GROUP BY c.id, c.name, c.icon, c.color
		ORDER BY SUM(t.amount) DESC
		LIMIT 5`, args...)
	if err != nil {
		return nil, fmt.Errorf("top categories: %w", err)
	}
	defer rows.Close()

	var cats []domain.CategorySummary
	for rows.Next() {
		var cs domain.CategorySummary
		if err := rows.Scan(&cs.ID, &cs.Name, &cs.Icon, &cs.Color, &cs.Count, &cs.Total); err != nil {
			return nil, fmt.Errorf("scan category summary: %w", err)
		}
		cats = append(cats, cs)
	}
	if cats == nil {
		cats = []domain.CategorySummary{}
	}
	return cats, rows.Err()
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
