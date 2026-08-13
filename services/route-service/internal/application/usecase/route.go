package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/service"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/in"
	"github.com/accessible-path/route-service/internal/ports/out"
	"github.com/google/uuid"
)

var ErrInvalidBBox = errors.New("invalid bbox, expected minLat,minLon,maxLat,maxLon")

const rerouteDistanceMeters = 25

type routeUseCase struct {
	router    *service.Router
	graphRepo out.GraphRepository
	routeRepo out.RouteRepository
	osmClient out.OSMProvider
	eventBus  out.EventBus
	barrierClient out.BarrierProvider
}

func NewRouteUseCase(router *service.Router, graphRepo out.GraphRepository, routeRepo out.RouteRepository, osmClient out.OSMProvider, eventBus out.EventBus, barrierClient out.BarrierProvider) in.RouteUseCase {
	return &routeUseCase{
		router:        router,
		graphRepo:     graphRepo,
		routeRepo:     routeRepo,
		osmClient:     osmClient,
		eventBus:      eventBus,
		barrierClient: barrierClient,
	}
}

func (u *routeUseCase) BuildRoute(start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.Route, error) {
	return u.router.BuildRoute(start, finish, profile)
}

func (u *routeUseCase) SaveRoute(userID uuid.UUID, start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.SavedRoute, error) {
	route, err := u.router.BuildRoute(start, finish, profile)
	if err != nil {
		return nil, err
	}

	saved := entity.NewSavedRoute(
		userID, route, profile,
		start.Latitude, start.Longitude,
		finish.Latitude, finish.Longitude,
	)
	if err := u.routeRepo.SaveRoute(saved); err != nil {
		return nil, err
	}

	return saved, nil
}

func (u *routeUseCase) ListRoutes(userID uuid.UUID) ([]*entity.SavedRoute, error) {
	return u.routeRepo.ListRoutesByUser(userID)
}

// HandleBarrierEvent rebuilds saved routes that pass within rerouteDistanceMeters
// of the barrier coordinates from barrier.approved / barrier.resolved events.
func (u *routeUseCase) HandleBarrierEvent(event map[string]interface{}) error {
	payload, _ := event["payload"].(map[string]interface{})
	if payload == nil {
		return nil
	}

	barrierID, _ := payload["barrier_id"].(string)
	if barrierID == "" {
		return nil
	}

	lat, lon := 0.0, 0.0
	coords, _ := payload["coordinates"].(map[string]interface{})
	if coords == nil {
		// moderation-service publishes barrier.approved without coordinates,
		// resolve the barrier via the barrier-service API.
		if u.barrierClient == nil {
			return nil
		}
		barrier, err := u.barrierClient.GetBarrier(barrierID)
		if err != nil {
			return err
		}
		if barrier == nil {
			return nil
		}
		lat, lon = barrier.Latitude, barrier.Longitude
	} else {
		if latVal, ok := coords["lat"].(float64); ok {
			lat = latVal
		}
		if lonVal, ok := coords["lon"].(float64); ok {
			lon = lonVal
		}
	}

	allRoutes, err := u.routeRepo.ListAllRoutes()
	if err != nil {
		return err
	}

	for _, saved := range allRoutes {
		if !routePassesNear(saved.Points, lat, lon, rerouteDistanceMeters) {
			continue
		}

		rebuilt, err := u.router.BuildRoute(
			valueobject.Coordinates{Latitude: saved.StartLat, Longitude: saved.StartLon},
			valueobject.Coordinates{Latitude: saved.FinishLat, Longitude: saved.FinishLon},
			saved.MobilityProfile,
		)
		if err != nil {
			continue
		}

		if !routeChanged(saved.Points, rebuilt.Nodes) {
			continue
		}

		saved.Update(rebuilt)
		if err := u.routeRepo.UpdateRoute(saved); err != nil {
			continue
		}

		if u.eventBus != nil {
			_ = u.eventBus.Publish("route.updated", map[string]interface{}{
				"event_type":   "route.updated",
				"aggregate_id": saved.ID.String(),
				"payload": map[string]interface{}{
					"route_id":   saved.ID.String(),
					"user_id":    saved.UserID.String(),
					"barrier_id": barrierID,
					"reason":     "barrier_on_route",
				},
			})
		}
	}

	return nil
}

// routePassesNear reports whether the polyline passes within maxDistMeters of (lat, lon).
func routePassesNear(points []*entity.Node, lat, lon float64, maxDistMeters float64) bool {
	if len(points) == 0 {
		return false
	}
	for i := 0; i < len(points)-1; i++ {
		a := valueobject.Coordinates{Latitude: points[i].Latitude, Longitude: points[i].Longitude}
		b := valueobject.Coordinates{Latitude: points[i+1].Latitude, Longitude: points[i+1].Longitude}
		d := pointToSegmentMeters(valueobject.Coordinates{Latitude: lat, Longitude: lon}, a, b)
		if d <= maxDistMeters {
			return true
		}
	}
	return false
}

// pointToSegmentMeters returns the distance in meters from p to segment [a, b].
func pointToSegmentMeters(p, a, b valueobject.Coordinates) float64 {
	// Convert to a local equirectangular projection in meters.
	midLat := (a.Latitude + b.Latitude) / 2 * math.Pi / 180
	mPerDegLat := 111132.0
	mPerDegLon := 111132.0 * math.Cos(midLat)

	px := p.Longitude * mPerDegLon
	py := p.Latitude * mPerDegLat
	ax := a.Longitude * mPerDegLon
	ay := a.Latitude * mPerDegLat
	bx := b.Longitude * mPerDegLon
	by := b.Latitude * mPerDegLat

	dx := bx - ax
	dy := by - ay
	if dx == 0 && dy == 0 {
		return math.Hypot(px-ax, py-ay)
	}

	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))

	cx := ax + t*dx
	cy := ay + t*dy
	return math.Hypot(px-cx, py-cy)
}

// routeChanged compares two polylines by their node IDs.
func routeChanged(prev, next []*entity.Node) bool {
	if len(prev) != len(next) {
		return true
	}
	for i := range prev {
		if prev[i].ID != next[i].ID {
			return true
		}
	}
	return false
}

func (u *routeUseCase) ImportOSM(bbox string) error {
	minLat, minLon, maxLat, maxLon, err := parseBBox(bbox)
	if err != nil {
		return err
	}

	graph, err := u.osmClient.FetchRoads(context.Background(), minLat, minLon, maxLat, maxLon)
	if err != nil {
		return fmt.Errorf("fetch OSM roads: %w", err)
	}

	if err := u.graphRepo.SaveGraph(graph); err != nil {
		return fmt.Errorf("save graph: %w", err)
	}

	u.router.SetGraph(graph)
	return nil
}

func parseBBox(bbox string) (float64, float64, float64, float64, error) {
	parts := strings.Split(bbox, ",")
	if len(parts) != 4 {
		return 0, 0, 0, 0, ErrInvalidBBox
	}

	vals := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, 0, 0, 0, ErrInvalidBBox
		}
		vals[i] = v
	}

	return vals[0], vals[1], vals[2], vals[3], nil
}