package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapiv1 "github.com/accessible-path/notification-service/internal/adapters/in/http"
	"github.com/accessible-path/notification-service/internal/adapters/out/eventbus"
	"github.com/accessible-path/notification-service/internal/adapters/out/postgres"
	"github.com/accessible-path/notification-service/internal/application/usecase"
	"github.com/accessible-path/notification-service/internal/domain/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := loadConfig()

	pool, err := pgxpool.New(context.Background(), cfg.databaseURL())
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

	notificationRepo := postgres.NewPostgresNotificationRepository(pool)
	
	redisEventBus := eventbus.NewRedisEventBus(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	wsManager := eventbus.NewWebSocketManager()
	go wsManager.Start()

	notificationService := service.NewNotificationService(notificationRepo, redisEventBus, wsManager)
	notificationUseCase := usecase.NewNotificationUseCase(notificationService)
	notificationHandler := httpapiv1.NewNotificationHandler(notificationUseCase, wsManager)

	go func() {
		if err := redisEventBus.Subscribe("barrier.approved", "notification-group", "notification-consumer-1", func(event map[string]interface{}) {
			if err := notificationService.HandleBarrierApproved(event); err != nil {
				logger.Error("handle barrier.approved", zap.Error(err))
			}
		}); err != nil {
			logger.Error("barrier.approved consumer error", zap.Error(err))
		}
	}()

	go func() {
		if err := redisEventBus.Subscribe("barrier.resolved", "notification-group", "notification-consumer-2", func(event map[string]interface{}) {
			if err := notificationService.HandleBarrierResolved(event); err != nil {
				logger.Error("handle barrier.resolved", zap.Error(err))
			}
		}); err != nil {
			logger.Error("barrier.resolved consumer error", zap.Error(err))
		}
	}()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	api := e.Group("/api/v1", httpapiv1.JWTUserIDMiddleware(cfg.JWTPublicKeyPath))
	notificationHandler.RegisterRoutes(api.Group("/notifications"))

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
	Port             string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	JWTPublicKeyPath string
}

func loadConfig() config {
	return config{
		Port:             getEnv("PORT", "8083"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "notification_user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "notification_pass"),
		PostgresDB:       getEnv("POSTGRES_DB", "notification_db"),
		RedisAddr:        getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          0,
		JWTPublicKeyPath: getEnv("JWT_PUBLIC_KEY_PATH", "/keys/public.pem"),
	}
}

func (c config) databaseURL() string {
	return "postgres://" + c.PostgresUser + ":" + c.PostgresPassword + "@" + c.PostgresHost + ":" + c.PostgresPort + "/" + c.PostgresDB + "?sslmode=disable"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}