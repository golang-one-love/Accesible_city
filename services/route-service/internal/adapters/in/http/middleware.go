package http

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var jwtPublicKey interface{}

func JWTUserIDMiddleware() echo.MiddlewareFunc {
	if jwtPublicKey == nil {
		key, err := loadPublicKey(getEnv("JWT_PUBLIC_KEY_PATH", "/keys/public.pem"))
		if err != nil {
			panic(fmt.Sprintf("load public key: %v", err))
		}
		jwtPublicKey = key
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := extractToken(c)
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
			}

			claims, err := validateAccessToken(tokenString, jwtPublicKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			}

			userID, err := uuid.Parse(claims["sub"].(string))
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token subject"})
			}

			c.Set("user_id", userID.String())
			return next(c)
		}
	}
}

func extractToken(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	if token := c.QueryParam("token"); token != "" {
		return token
	}

	return ""
}

func loadPublicKey(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(data)
}

func validateAccessToken(tokenString string, publicKey interface{}) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims["type"] != "access" {
		return nil, fmt.Errorf("not an access token")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil, fmt.Errorf("missing subject")
	}

	return claims, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}