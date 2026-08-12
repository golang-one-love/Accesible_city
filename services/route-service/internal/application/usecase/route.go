package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/service"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/in"
	"github.com/accessible-path/route-service/internal/ports/out"
)

var ErrInvalidBBox = errors.New("invalid bbox, expected minLat,minLon,maxLat,maxLon")

type routeUseCase struct {
	router    *service.Router
	graphRepo out.GraphRepository
	osmClient out.OSMProvider
}

func NewRouteUseCase(router *service.Router, graphRepo out.GraphRepository, osmClient out.OSMProvider) in.RouteUseCase {
	return &routeUseCase{
		router:    router,
		graphRepo: graphRepo,
		osmClient: osmClient,
	}
}

func (u *routeUseCase) BuildRoute(start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.Route, error) {
	return u.router.BuildRoute(start, finish, profile)
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