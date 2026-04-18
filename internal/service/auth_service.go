package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/antonidev/dompet-santuy/internal/domain"
	"github.com/antonidev/dompet-santuy/internal/repository"
	"github.com/antonidev/dompet-santuy/internal/util"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	tokenRepo   *repository.RefreshTokenRepository
	jwtManager  *util.JWTManager
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.RefreshTokenRepository,
	jwtManager *util.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, repository.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, user.ID, user.Email)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*domain.AuthResponse, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(rawRefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenHash := util.HashToken(rawRefreshToken)
	stored, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrTokenRevoked
	}
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}

	if time.Now().UTC().After(stored.ExpiresAt) {
		_ = s.tokenRepo.DeleteByHash(ctx, tokenHash)
		return nil, ErrInvalidToken
	}

	// Rotate: delete old, issue new (refresh token rotation)
	if err := s.tokenRepo.DeleteByHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	return s.issueTokenPair(ctx, claims.UserID, claims.Email)
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	tokenHash := util.HashToken(rawRefreshToken)
	if err := s.tokenRepo.DeleteByHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	if err := s.tokenRepo.DeleteAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("revoke all tokens: %w", err)
	}
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID, email string) (*domain.AuthResponse, error) {
	pair, err := s.jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	rt := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: util.HashToken(pair.RefreshToken),
		ExpiresAt: time.Now().UTC().Add(s.jwtManager.RefreshExpiryDuration()),
	}
	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(time.Until(pair.AccessExpiry).Seconds()),
	}, nil
}
