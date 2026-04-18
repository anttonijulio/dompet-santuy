package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	DSN      string
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpiryMinutes int
	RefreshExpiryDays   int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessExpiry, err := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY_MINUTES: %w", err)
	}
	refreshExpiry, err := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY_DAYS: %w", err)
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "dompet_santuy")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	cfg := &Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			DSN:      dsn,
		},
		JWT: JWTConfig{
			AccessSecret:        getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret:       getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpiryMinutes: accessExpiry,
			RefreshExpiryDays:   refreshExpiry,
		},
	}

	if cfg.JWT.AccessSecret == "" || cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
