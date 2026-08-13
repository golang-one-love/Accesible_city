package service

import (
	"errors"
	"testing"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/out"
)

type fakeBarrierProvider struct {
	barriers []out.BarrierInfo
	err      error
	calls    int
}

func (f *fakeBarrierProvider) GetBarriersInBounds(minLat, minLon, maxLat, maxLon float64) ([]out.BarrierInfo, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.barriers, nil
}

func (f *fakeBarrierProvider) GetBarrier(barrierID string) (*out.BarrierInfo, error) {
	return nil, nil
}

// buildTestGrid builds a gridGraph x gridGraph mesh of nodes spaced 0.001 deg
// apart, connected both ways along grid lines.
func buildTestGrid(gridGraph int) *entity.Graph {
	g := entity.NewGraph()
	for i := 0; i < gridGraph; i++ {
		for j := 0; j < gridGraph; j++ {
			id := nodeID(i, j)
			g.AddNode(&entity.Node{
				ID:        id,
				Latitude:  55.7 + float64(i)*0.001,
				Longitude: 37.5 + float64(j)*0.001,
			})
		}
	}
	for i := 0; i < gridGraph; i++ {
		for j := 0; j < gridGraph; j++ {
			id := nodeID(i, j)
			if i+1 < gridGraph {
				addBidirectionalEdge(g, id, nodeID(i+1, j), 111)
			}
			if j+1 < gridGraph {
				addBidirectionalEdge(g, id, nodeID(i, j+1), 111)
			}
		}
	}
	g.BuildIndex()
	return g
}

func nodeID(i, j int) string {
	return "n_" + string(rune('0'+i)) + string(rune('0'+j))
}

func addBidirectionalEdge(g *entity.Graph, from, to string, distance float64) {
	g.AddEdge(&entity.Edge{From: from, To: to, Distance: distance, Severity: 1})
	g.AddEdge(&entity.Edge{From: to, To: from, Distance: distance, Severity: 1})
}

func newTestRouter() *Router {
	return NewRouter(buildTestGrid(20), &fakeBarrierProvider{})
}

func TestBuildRouteLongDistance(t *testing.T) {
	r := newTestRouter()

	start := valueobject.Coordinates{Latitude: 55.7005, Longitude: 37.5005}
	finish := valueobject.Coordinates{Latitude: 55.7195, Longitude: 37.5195}

	route, err := r.BuildRoute(start, finish, entity.MobilityDefault)
	if err != nil {
		t.Fatalf("expected a long route to build, got error: %v", err)
	}
	if len(route.Nodes) < 2 {
		t.Fatalf("expected a multi-node route, got %d nodes", len(route.Nodes))
	}
	if route.TotalDistance <= 0 {
		t.Fatalf("expected positive distance, got %f", route.TotalDistance)
	}
}

func TestBuildRouteStartOutOfCoverage(t *testing.T) {
	r := newTestRouter()

	start := valueobject.Coordinates{Latitude: 54.0, Longitude: 37.5}
	finish := valueobject.Coordinates{Latitude: 55.7105, Longitude: 37.5105}

	_, err := r.BuildRoute(start, finish, entity.MobilityDefault)
	if !errors.Is(err, ErrStartOutOfCoverage) {
		t.Fatalf("expected ErrStartOutOfCoverage, got %v", err)
	}
}

func TestBuildRouteFinishOutOfCoverage(t *testing.T) {
	r := newTestRouter()

	start := valueobject.Coordinates{Latitude: 55.7105, Longitude: 37.5105}
	finish := valueobject.Coordinates{Latitude: 56.9, Longitude: 37.6}

	_, err := r.BuildRoute(start, finish, entity.MobilityDefault)
	if !errors.Is(err, ErrFinishOutOfCoverage) {
		t.Fatalf("expected ErrFinishOutOfCoverage, got %v", err)
	}
}

func TestBuildRouteNoPath(t *testing.T) {
	g := entity.NewGraph()
	g.AddNode(&entity.Node{ID: "a", Latitude: 55.70, Longitude: 37.50})
	g.AddNode(&entity.Node{ID: "b", Latitude: 55.71, Longitude: 37.51})
	g.AddEdge(&entity.Edge{From: "a", To: "b", Distance: 100, Severity: 1})
	g.BuildIndex()

	r := NewRouter(g, &fakeBarrierProvider{})

	start := valueobject.Coordinates{Latitude: 55.7002, Longitude: 37.5002}
	finish := valueobject.Coordinates{Latitude: 56.0, Longitude: 37.5}

	_, err := r.BuildRoute(start, finish, entity.MobilityDefault)
	if !errors.Is(err, ErrFinishOutOfCoverage) {
		t.Fatalf("expected coverage error, got %v", err)
	}
}

func TestBuildRouteBarrierFailureIsSoft(t *testing.T) {
	g := buildTestGrid(20)
	provider := &fakeBarrierProvider{err: errors.New("barrier service returned 500")}
	r := NewRouter(g, provider)

	start := valueobject.Coordinates{Latitude: 55.7005, Longitude: 37.5005}
	finish := valueobject.Coordinates{Latitude: 55.7105, Longitude: 37.5105}

	route, err := r.BuildRoute(start, finish, entity.MobilityWheelchair)
	if err != nil {
		t.Fatalf("barrier failure should not break routing, got: %v", err)
	}
	if route == nil || len(route.Nodes) < 2 {
		t.Fatalf("expected a route despite barrier failure")
	}
	if provider.calls == 0 {
		t.Fatalf("expected the barrier provider to be called")
	}
}

func TestBuildRouteBlocksObstacleNode(t *testing.T) {
	g := buildTestGrid(20)
	provider := &fakeBarrierProvider{barriers: []out.BarrierInfo{
		{ID: "b1", Latitude: 55.7055, Longitude: 37.5055, Severity: 1},
	}}
	r := NewRouter(g, provider)

	start := valueobject.Coordinates{Latitude: 55.7005, Longitude: 37.5005}
	finish := valueobject.Coordinates{Latitude: 55.7105, Longitude: 37.5105}

	route, err := r.BuildRoute(start, finish, entity.MobilityWheelchair)
	if err != nil {
		t.Fatalf("expected an alternative path around the blocked node, got: %v", err)
	}
	if len(route.Nodes) < 2 {
		t.Fatalf("expected a multi-node route")
	}
}

func TestBuildRouteBlockedUnreachable(t *testing.T) {
	g := entity.NewGraph()
	g.AddNode(&entity.Node{ID: "a", Latitude: 55.70, Longitude: 37.50})
	g.AddNode(&entity.Node{ID: "b", Latitude: 55.71, Longitude: 37.51})
	g.AddEdge(&entity.Edge{From: "a", To: "b", Distance: 100, Severity: 1})
	g.AddEdge(&entity.Edge{From: "b", To: "a", Distance: 100, Severity: 1})
	g.BuildIndex()

	provider := &fakeBarrierProvider{barriers: []out.BarrierInfo{
		{ID: "b1", Latitude: 55.71, Longitude: 37.51, Severity: 1},
	}}
	r := NewRouter(g, provider)

	start := valueobject.Coordinates{Latitude: 55.7002, Longitude: 37.5002}
	finish := valueobject.Coordinates{Latitude: 55.7102, Longitude: 37.5102}

	_, err := r.BuildRoute(start, finish, entity.MobilityWheelchair)
	if !errors.Is(err, ErrNoPathFound) {
		t.Fatalf("expected ErrNoPathFound, got %v", err)
	}
}
