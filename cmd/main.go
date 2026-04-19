package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antonidev/dompet-santuy/internal/config"
	"github.com/antonidev/dompet-santuy/internal/database"
	"github.com/antonidev/dompet-santuy/internal/handler"
	appmw "github.com/antonidev/dompet-santuy/internal/middleware"
	"github.com/antonidev/dompet-santuy/internal/repository"
	"github.com/antonidev/dompet-santuy/internal/response"
	"github.com/antonidev/dompet-santuy/internal/service"
	"github.com/antonidev/dompet-santuy/internal/util"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewMySQL(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	jwtManager := util.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiryMinutes,
		cfg.JWT.RefreshExpiryDays,
	)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authSvc := service.NewAuthService(userRepo, tokenRepo, jwtManager)
	categorySvc := service.NewCategoryService(categoryRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, categoryRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	transactionHandler := handler.NewTransactionHandler(transactionSvc)

	// Cleanup expired tokens every hour
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := tokenRepo.DeleteExpired(context.Background()); err != nil {
				log.Printf("cleanup expired tokens: %v", err)
			}
		}
	}()

	e := echo.New()
	e.HideBanner = true
	e.Validator = util.NewValidator()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		message := "internal server error"
		var validationErrors []string

		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
			switch v := he.Message.(type) {
			case map[string]any:
				if msg, ok := v["message"].(string); ok {
					message = msg
				}
				if errs, ok := v["errors"].([]string); ok {
					validationErrors = errs
				}
			case string:
				message = v
			default:
				message = fmt.Sprintf("%v", v)
			}
		}

		c.JSON(code, response.Response[any]{ //nolint:errcheck
			Success: false,
			Message: message,
			Errors:  validationErrors,
		})
	}

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Routes
	v1 := e.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout)

	protected := v1.Group("")
	protected.Use(appmw.JWTAuth(jwtManager))
	protected.GET("/me", authHandler.Me)
	protected.POST("/auth/logout-all", authHandler.LogoutAll)

	protected.GET("/categories", categoryHandler.List)
	protected.POST("/categories", categoryHandler.Create)
	protected.GET("/categories/:id", categoryHandler.Get)
	protected.PUT("/categories/:id", categoryHandler.Update)
	protected.DELETE("/categories/:id", categoryHandler.Delete)
	protected.GET("/transactions", transactionHandler.List)
	protected.POST("/transactions", transactionHandler.Create)
	protected.GET("/transactions/summary", transactionHandler.Summary)
	protected.GET("/transactions/:id", transactionHandler.Get)
	protected.PUT("/transactions/:id", transactionHandler.Update)
	protected.DELETE("/transactions/:id", transactionHandler.Delete)

	v1.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf(":%s", cfg.App.Port)
		log.Printf("starting server on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
