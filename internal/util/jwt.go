package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessExpiry time.Time
}

type JWTManager struct {
	accessSecret        []byte
	refreshSecret       []byte
	accessExpiryMinutes int
	refreshExpiryDays   int
}

func NewJWTManager(accessSecret, refreshSecret string, accessMinutes, refreshDays int) *JWTManager {
	return &JWTManager{
		accessSecret:        []byte(accessSecret),
		refreshSecret:       []byte(refreshSecret),
		accessExpiryMinutes: accessMinutes,
		refreshExpiryDays:   refreshDays,
	}
}

func (m *JWTManager) GenerateTokenPair(userID, email string) (*TokenPair, error) {
	accessExpiry := time.Now().UTC().Add(time.Duration(m.accessExpiryMinutes) * time.Minute)
	accessToken, err := m.generateToken(userID, email, accessExpiry, m.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshExpiry := time.Now().UTC().Add(time.Duration(m.refreshExpiryDays) * 24 * time.Hour)
	refreshToken, err := m.generateToken(userID, email, refreshExpiry, m.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExpiry: accessExpiry,
	}, nil
}

func (m *JWTManager) generateToken(userID, email string, expiry time.Time, secret []byte) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *JWTManager) ValidateAccessToken(tokenStr string) (*JWTClaims, error) {
	return m.validateToken(tokenStr, m.accessSecret)
}

func (m *JWTManager) ValidateRefreshToken(tokenStr string) (*JWTClaims, error) {
	return m.validateToken(tokenStr, m.refreshSecret)
}

func (m *JWTManager) validateToken(tokenStr string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func (m *JWTManager) RefreshExpiryDuration() time.Duration {
	return time.Duration(m.refreshExpiryDays) * 24 * time.Hour
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
