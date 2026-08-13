package http

import (
	"net/http"
	"strings"

	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/ports/in"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	authUseCase in.AuthUseCase
}

func NewAuthMiddleware(authUseCase in.AuthUseCase) *AuthMiddleware {
	return &AuthMiddleware{authUseCase: authUseCase}
}

func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header format"})
			}

			claims, err := m.authUseCase.ValidateAccessToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			}

			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
			return next(c)
		}
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...entity.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole, ok := c.Get("user_role").(entity.Role)
			if !ok {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			}

			for _, r := range roles {
				if userRole == r {
					return next(c)
				}
			}
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "insufficient permissions"})
		}
	}
}

func GetUserID(c echo.Context) string {
	return c.Get("user_id").(string)
}

func GetUserRole(c echo.Context) entity.Role {
	return c.Get("user_role").(entity.Role)
}