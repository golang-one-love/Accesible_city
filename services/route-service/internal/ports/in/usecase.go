package in

import (
	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
)

type RouteUseCase interface {
	BuildRoute(start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.Route, error)
	ImportOSM(bbox string) error
}