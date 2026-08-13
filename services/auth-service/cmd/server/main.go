package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapiv1 "github.com/accessible-path/auth-service/internal/adapters/in/http"
	"github.com/accessible-path/auth-service/internal/adapters/out/jwt"
	"github.com/accessible-path/auth-service/internal/adapters/out/postgres"
	"github.com/accessible-path/auth-service/internal/application/usecase"
	"github.com/accessible-path/auth-service/internal/domain/entity"
	"github.com/accessible-path/auth-service/internal/domain/service"
	"github.com/accessible-path/auth-service/internal/domain/valueobject"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := loadConfig()

	poolConfig, err := pgxpool.ParseConfig(cfg.databaseURL())
	if err != nil {
		logger.Fatal("failed to parse database config", zap.Error(err))
	}
	poolConfig.MaxConns = 3

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	if err := postgres.EnsureSchema(context.Background(), pool); err != nil {
		logger.Fatal("failed to ensure schema", zap.Error(err))
	}

	tokenGen, err := jwt.NewJWTTokenGenerator(
		cfg.JWTPrivateKeyPath,
		cfg.JWTPublicKeyPath,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	if err != nil {
		logger.Fatal("failed to create token generator", zap.Error(err))
	}

	userRepo := postgres.NewUserRepository(pool)
	roleAppRepo := postgres.NewRoleApplicationRepository(pool)

	if err := seedModerator(userRepo, logger, cfg); err != nil {
		logger.Fatal("failed to seed moderator", zap.Error(err))
	}

	authService := service.NewAuthService(
		userRepo,
		roleAppRepo,
		tokenGen,
		cfg.BCryptCost,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)

	authUseCase := usecase.NewAuthUseCase(authService)
	authMiddleware := httpapiv1.NewAuthMiddleware(authUseCase)
	authHandler := httpapiv1.NewAuthHandler(authUseCase, authMiddleware)

	e := echo.New()
	e.Validator = newValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	api := e.Group("/api/v1")
	authHandler.RegisterRoutes(api.Group("/auth"))

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal("forced shutdown", zap.Error(err))
	}
}

type config struct {
	Port                   string
	PostgresHost           string
	PostgresPort           string
	PostgresUser           string
	PostgresPassword       string
	PostgresDB             string
	PostgresSSLMode        string
	JWTPrivateKeyPath      string
	JWTPublicKeyPath       string
	JWTAccessTTL           time.Duration
	JWTRefreshTTL          time.Duration
	BCryptCost             int
	SeedModeratorEmail     string
	SeedModeratorPassword  string
}

func loadConfig() config {
	accessTTL, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))

	return config{
		Port:                  getEnv("PORT", "8081"),
		PostgresHost:          getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:          getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:          getEnv("POSTGRES_USER", "auth_user"),
		PostgresPassword:      getEnv("POSTGRES_PASSWORD", "auth_pass"),
		PostgresDB:            getEnv("POSTGRES_DB", "auth_db"),
		PostgresSSLMode:       getEnv("POSTGRES_SSLMODE", "disable"),
		JWTPrivateKeyPath:     getEnv("JWT_PRIVATE_KEY_PATH", "/keys/private.pem"),
		JWTPublicKeyPath:      getEnv("JWT_PUBLIC_KEY_PATH", "/keys/public.pem"),
		JWTAccessTTL:          accessTTL,
		JWTRefreshTTL:         refreshTTL,
		BCryptCost:            12,
		SeedModeratorEmail:    getEnv("SEED_MODERATOR_EMAIL", "moderator@accessible.city"),
		SeedModeratorPassword: getEnv("SEED_MODERATOR_PASSWORD", "Moderator2026!"),
	}
}

func (c config) databaseURL() string {
	return "postgres://" + c.PostgresUser + ":" + c.PostgresPassword + "@" + c.PostgresHost + ":" + c.PostgresPort + "/" + c.PostgresDB + "?sslmode=" + c.PostgresSSLMode
}

func seedModerator(userRepo *postgres.UserRepository, logger *zap.Logger, cfg config) error {
	email, err := valueobject.NewEmail(cfg.SeedModeratorEmail)
	if err != nil {
		return fmt.Errorf("invalid seed moderator email: %w", err)
	}

	_, err = userRepo.GetByEmail(email)
	if err == nil {
		return nil
	}
	if !errors.Is(err, service.ErrUserNotFound) {
		return err
	}

	password, err := valueobject.NewPassword(cfg.SeedModeratorPassword)
	if err != nil {
		return fmt.Errorf("invalid seed moderator password: %w", err)
	}

	hash, err := password.Hash(cfg.BCryptCost)
	if err != nil {
		return fmt.Errorf("hash seed moderator password: %w", err)
	}

	user := entity.NewUser(email, hash.String(), entity.RoleModerator)
	user.UpdateNickname("Модератор")

	if err := userRepo.Create(user); err != nil {
		return fmt.Errorf("create seed moderator: %w", err)
	}

	logger.Info("seeded moderator user", zap.String("email", email.String()))
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type validator struct{}

func newValidator() *validator {
	return &validator{}
}

func (v *validator) Validate(i interface{}) error {
	return nil
}