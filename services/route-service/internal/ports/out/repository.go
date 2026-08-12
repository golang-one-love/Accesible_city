package out

import (
	"context"

	"github.com/accessible-path/route-service/internal/domain/entity"
)

type GraphRepository interface {
	SaveGraph(graph *entity.Graph) error
	LoadGraph() (*entity.Graph, error)
}

type OSMProvider interface {
	FetchRoads(ctx context.Context, minLat, minLon, maxLat, maxLon float64) (*entity.Graph, error)
}

type BarrierProvider interface {
	GetBarriersInBounds(minLat, minLon, maxLat, maxLon float64) ([]BarrierInfo, error)
}

type BarrierInfo struct {
	ID       string
	Latitude  float64
	Longitude float64
	Severity  int
}