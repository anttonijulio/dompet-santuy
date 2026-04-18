package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("record not found")
var ErrDuplicateEmail = errors.New("email already registered")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.NewString()
	query := `INSERT INTO users (id, name, email, password) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		if isDuplicateEntry(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ?`
	row := r.db.QueryRowContext(ctx, query, email)

	var u domain.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var u domain.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Save(ctx context.Context, rt *domain.RefreshToken) error {
	rt.ID = uuid.NewString()
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, rt.ID, rt.UserID, rt.TokenHash, rt.ExpiresAt)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = ?`
	row := r.db.QueryRowContext(ctx, query, hash)

	var rt domain.RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) DeleteByHash(ctx context.Context, hash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = ?`
	_, err := r.db.ExecContext(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete all refresh tokens: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < ?`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("delete expired tokens: %w", err)
	}
	return nil
}

func isDuplicateEntry(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062"))
}
