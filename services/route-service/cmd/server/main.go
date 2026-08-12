package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapiv1 "github.com/accessible-path/route-service/internal/adapters/in/http"
	"github.com/accessible-path/route-service/internal/adapters/out/cache"
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

	graphCache := cache.NewMemoryGraphCache()
	graphRepo := postgres.NewPostgresGraphRepository(pool)

	osmClient := external.NewOSMClient(cfg.OSMOverpassURL)
	barrierClient := external.NewBarrierClient(cfg.BarrierServiceURL)

	router := service.NewRouter(entity.NewGraph(), barrierClient)
	routeUseCase := usecase.NewRouteUseCase(router, graphRepo, osmClient)

	graph, err := graphRepo.LoadGraph()
	if err != nil || len(graph.Nodes) == 0 {
		logger.Warn("no cached graph found, importing OSM roads...", zap.Error(err))
		if impErr := routeUseCase.ImportOSM(cfg.OSMDefaultBBox); impErr != nil {
			logger.Warn("OSM import failed, falling back to default grid", zap.Error(impErr))
			graph = createDefaultGrid()
		} else {
			graph, err = graphRepo.LoadGraph()
			if err != nil || len(graph.Nodes) == 0 {
				logger.Warn("failed to reload imported graph, using default grid", zap.Error(err))
				graph = createDefaultGrid()
			}
		}
	}
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
	Port             string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	BarrierServiceURL string
	OSMOverpassURL   string
	OSMDefaultBBox   string
}

func loadConfig() config {
	return config{
		Port:              getEnv("PORT", "8082"),
		PostgresHost:      getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:      getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:      getEnv("POSTGRES_USER", "route_user"),
		PostgresPassword:  getEnv("POSTGRES_PASSWORD", "route_pass"),
		PostgresDB:        getEnv("POSTGRES_DB", "route_db"),
		BarrierServiceURL: getEnv("BARRIER_SERVICE_URL", "http://barrier-service:8000"),
		OSMOverpassURL:    getEnv("OSM_OVERPASS_URL", "https://overpass-api.de/api/interpreter"),
		OSMDefaultBBox:    getEnv("OSM_DEFAULT_BBOX", "55.7000,37.5200,55.8000,37.7200"),
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
			nodeID := fmt.Sprintf("node_%d_%d", i, j)
			lat := baseLat + float64(i)*spacing
			lon := baseLon + float64(j)*spacing
			
			node := &entity.Node{
				ID:        nodeID,
				Latitude:  lat,
				Longitude: lon,
			}
			graph.AddNode(node)
		}
	}

	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			nodeID := fmt.Sprintf("node_%d_%d", i, j)
			
			if i > 0 {
				neighborID := fmt.Sprintf("node_%d_%d", i-1, j)
				graph.AddEdge(&entity.Edge{
					From:     nodeID,
					To:       neighborID,
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
			if i < gridSize-1 {
				neighborID := fmt.Sprintf("node_%d_%d", i+1, j)
				graph.AddEdge(&entity.Edge{
					From:     nodeID,
					To:       neighborID,
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
			if j > 0 {
				neighborID := fmt.Sprintf("node_%d_%d", i, j-1)
				graph.AddEdge(&entity.Edge{
					From:     nodeID,
					To:       neighborID,
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
			if j < gridSize-1 {
				neighborID := fmt.Sprintf("node_%d_%d", i, j+1)
				graph.AddEdge(&entity.Edge{
					From:     nodeID,
					To:       neighborID,
					Distance: spacing * 111000,
					Severity: 1,
				})
			}
		}
	}

	return graph
}