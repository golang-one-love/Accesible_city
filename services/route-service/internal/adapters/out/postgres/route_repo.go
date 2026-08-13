package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func parseNodeIDLegacy(key string) int64 {
	key = strings.TrimPrefix(key, "osm:")
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

type PostgresRouteRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRouteRepository(pool *pgxpool.Pool) *PostgresRouteRepository {
	return &PostgresRouteRepository{pool: pool}
}

func (r *PostgresRouteRepository) SaveRoute(route *entity.SavedRoute) error {
	ctx := context.Background()

	pointsJSON, err := json.Marshal(route.Points)
	if err != nil {
		return fmt.Errorf("marshal points: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO routes (id, user_id, start_lat, start_lon, finish_lat, finish_lon,
			mobility_profile, points, total_distance, max_severity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		route.ID, route.UserID,
		route.StartLat, route.StartLon, route.FinishLat, route.FinishLon,
		route.MobilityProfile, pointsJSON, route.TotalDistance, route.MaxSeverity,
		route.CreatedAt, route.UpdatedAt,
	)

	return err
}

func (r *PostgresRouteRepository) ListRoutesByUser(userID uuid.UUID) ([]*entity.SavedRoute, error) {
	ctx := context.Background()

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, start_lat, start_lon, finish_lat, finish_lon,
			mobility_profile, points, total_distance, max_severity, created_at, updated_at
		FROM routes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := make([]*entity.SavedRoute, 0)
	for rows.Next() {
		route, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, rows.Err()
}

func (r *PostgresRouteRepository) ListAllRoutes() ([]*entity.SavedRoute, error) {
	ctx := context.Background()

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, start_lat, start_lon, finish_lat, finish_lon,
			mobility_profile, points, total_distance, max_severity, created_at, updated_at
		FROM routes
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := make([]*entity.SavedRoute, 0)
	for rows.Next() {
		route, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, rows.Err()
}

func (r *PostgresRouteRepository) UpdateRoute(route *entity.SavedRoute) error {
	ctx := context.Background()

	pointsJSON, err := json.Marshal(route.Points)
	if err != nil {
		return fmt.Errorf("marshal points: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE routes
		SET points = $2, total_distance = $3, max_severity = $4, updated_at = $5
		WHERE id = $1
	`, route.ID, pointsJSON, route.TotalDistance, route.MaxSeverity, route.UpdatedAt)

	return err
}

type routeScanner interface {
	Scan(dest ...interface{}) error
}

func scanRoute(row routeScanner) (*entity.SavedRoute, error) {
	var route entity.SavedRoute
	var pointsJSON []byte

	err := row.Scan(
		&route.ID, &route.UserID,
		&route.StartLat, &route.StartLon, &route.FinishLat, &route.FinishLon,
		&route.MobilityProfile, &pointsJSON, &route.TotalDistance, &route.MaxSeverity,
		&route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(pointsJSON, &route.Points); err != nil {
		// Legacy points stored string ids ("osm:<id>"); convert them so old
		// saved routes keep working with the int64 node ids.
		var legacy []struct {
			ID        string
			Latitude  float64
			Longitude float64
		}
		if lerr := json.Unmarshal(pointsJSON, &legacy); lerr != nil {
			return nil, fmt.Errorf("unmarshal points: %w", err)
		}
		route.Points = make([]*entity.Node, len(legacy))
		for i, p := range legacy {
			route.Points[i] = &entity.Node{
				ID:        parseNodeIDLegacy(p.ID),
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
			}
		}
	}

	return &route, nil
}
