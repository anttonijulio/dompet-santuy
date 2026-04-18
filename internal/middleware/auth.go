package middleware

import (
	"net/http"
	"strings"

	"github.com/antonidev/dompet-santuy/internal/util"
	"github.com/labstack/echo/v4"
)

const UserIDKey = "user_id"
const UserEmailKey = "user_email"

func JWTAuth(jwtManager *util.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			claims, err := jwtManager.ValidateAccessToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set(UserIDKey, claims.UserID)
			c.Set(UserEmailKey, claims.Email)
			return next(c)
		}
	}
}
