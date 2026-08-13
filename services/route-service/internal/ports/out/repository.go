package out

import (
	"context"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/google/uuid"
)

type GraphRepository interface {
	SaveGraph(graph *entity.Graph) error
	LoadGraph() (*entity.Graph, error)
}

type RouteRepository interface {
	SaveRoute(route *entity.SavedRoute) error
	ListRoutesByUser(userID uuid.UUID) ([]*entity.SavedRoute, error)
	ListAllRoutes() ([]*entity.SavedRoute, error)
	UpdateRoute(route *entity.SavedRoute) error
}

type OSMProvider interface {
	FetchRoads(ctx context.Context, minLat, minLon, maxLat, maxLon float64) (*entity.Graph, error)
}

type EventBus interface {
	Publish(stream string, event map[string]interface{}) error
	Subscribe(stream, group, consumer string, handler func(map[string]interface{})) error
}

type BarrierProvider interface {
	GetBarriersInBounds(minLat, minLon, maxLat, maxLon float64) ([]BarrierInfo, error)
	GetBarrier(barrierID string) (*BarrierInfo, error)
}

type BarrierInfo struct {
	ID        string
	Latitude  float64
	Longitude float64
	Severity  int
}