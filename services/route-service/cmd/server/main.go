package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapiv1 "github.com/accessible-path/route-service/internal/adapters/in/http"
	"github.com/accessible-path/route-service/internal/adapters/out/cache"
	"github.com/accessible-path/route-service/internal/adapters/out/eventbus"
	"github.com/accessible-path/route-service/internal/adapters/out/external"
	"github.com/accessible-path/route-service/internal/adapters/out/postgres"
	"github.com/accessible-path/route-service/internal/application/usecase"
	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/service"
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

	waitForDB(logger, pool)

	if err := postgres.EnsureSchema(context.Background(), pool); err != nil {
		logger.Fatal("failed to ensure schema", zap.Error(err))
	}

	graphCache := cache.NewMemoryGraphCache()
	graphRepo := postgres.NewPostgresGraphRepository(pool)
	routeRepo := postgres.NewPostgresRouteRepository(pool)

	redisEventBus := eventbus.NewRedisEventBus(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	osmClient := external.NewOSMClient(cfg.OSMOverpassURL)
	barrierClient := external.NewBarrierClient(cfg.BarrierServiceURL)

	router := service.NewRouter(entity.NewGraph(), barrierClient)
	routeUseCase := usecase.NewRouteUseCase(router, graphRepo, routeRepo, osmClient, redisEventBus, barrierClient)

	go func() {
		if err := redisEventBus.Subscribe("barrier.approved", "route-group", "route-consumer-approved", func(event map[string]interface{}) {
			if err := routeUseCase.HandleBarrierEvent(event); err != nil {
				logger.Error("handle barrier.approved reroute", zap.Error(err))
			}
		}); err != nil {
			logger.Error("barrier.approved consumer error", zap.Error(err))
		}
	}()

	go func() {
		if err := redisEventBus.Subscribe("barrier.resolved", "route-group", "route-consumer-resolved", func(event map[string]interface{}) {
			if err := routeUseCase.HandleBarrierEvent(event); err != nil {
				logger.Error("handle barrier.resolved reroute", zap.Error(err))
			}
		}); err != nil {
			logger.Error("barrier.resolved consumer error", zap.Error(err))
		}
	}()

	graph, err := graphRepo.LoadGraph()
	if err != nil || graph.NumNodes() == 0 {
		logger.Warn("no cached graph found, importing OSM roads...", zap.Error(err))
		if impErr := routeUseCase.ImportOSM(cfg.OSMDefaultBBox); impErr != nil {
			logger.Warn("OSM import failed, falling back to default grid", zap.Error(impErr))
			graph = createDefaultGrid()
		} else {
			graph, err = graphRepo.LoadGraph()
			if err != nil || graph.NumNodes() == 0 {
				logger.Warn("failed to reload imported graph, using default grid", zap.Error(err))
				graph = createDefaultGrid()
			}
		}
	}
	graph.BuildIndex()
	router.SetGraph(graph)
	graphCache.Set(graph)

	routeHandler := httpapiv1.NewRouteHandler(routeUseCase)

	e := echo.New()
	e.Validator = newValidator()
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

	api := e.Group("/api/v1")
	routeHandler.RegisterRoutes(api.Group("/routes"))

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
	Port              string
	PostgresHost      string
	PostgresPort      string
	PostgresUser      string
	PostgresPassword  string
	PostgresDB        string
	PostgresSSLMode   string
	BarrierServiceURL string
	OSMOverpassURL    string
	OSMDefaultBBox    string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
}

func loadConfig() config {
	return config{
		Port:              getEnv("PORT", "8082"),
		PostgresHost:      getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:      getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:      getEnv("POSTGRES_USER", "route_user"),
		PostgresPassword:  getEnv("POSTGRES_PASSWORD", "route_pass"),
		PostgresDB:        getEnv("POSTGRES_DB", "route_db"),
		PostgresSSLMode:   getEnv("POSTGRES_SSLMODE", "disable"),
		BarrierServiceURL: getEnv("BARRIER_SERVICE_URL", "http://barrier-service:8000"),
		OSMOverpassURL:    getEnv("OSM_OVERPASS_URL", "https://overpass-api.de/api/interpreter"),
		OSMDefaultBBox:    getEnv("OSM_DEFAULT_BBOX", "55.7000,37.5200,55.8000,37.7200"),
		RedisAddr:         getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           0,
	}
}

func waitForDB(logger *zap.Logger, pool *pgxpool.Pool) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := pool.Ping(ctx)
		cancel()
		if err == nil {
			return
		}
		logger.Warn("database not ready, retrying in 5s", zap.Error(err))
		time.Sleep(5 * time.Second)
	}
}

func (c config) databaseURL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   c.PostgresHost + ":" + c.PostgresPort,
		Path:   "/" + c.PostgresDB,
	}
	q := u.Query()
	q.Set("sslmode", c.PostgresSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
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

func createDefaultGrid() *entity.Graph {
	graph := entity.NewGraph()

	gridSize := 10
	spacing := 0.001
	baseLat := 55.75
	baseLon := 37.61

	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			nodeID := int64(i*gridSize + j)
			lat := baseLat + float64(i)*spacing
			lon := baseLon + float64(j)*spacing

			graph.AddNode(entity.Node{
				ID:        nodeID,
				Latitude:  lat,
				Longitude: lon,
			})
		}
	}

	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			nodeID := int64(i*gridSize + j)

			if i > 0 {
				graph.AddEdge(nodeID, entity.Edge{
					To:       int64((i-1)*gridSize + j),
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
			if i < gridSize-1 {
				graph.AddEdge(nodeID, entity.Edge{
					To:       int64((i+1)*gridSize + j),
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
			if j > 0 {
				graph.AddEdge(nodeID, entity.Edge{
					To:       int64(i*gridSize + j - 1),
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
			if j < gridSize-1 {
				graph.AddEdge(nodeID, entity.Edge{
					To:       int64(i*gridSize + j + 1),
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
		}
	}

	return graph
}
