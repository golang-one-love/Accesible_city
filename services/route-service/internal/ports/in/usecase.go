package in

import (
	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/google/uuid"
)

type RouteUseCase interface {
	BuildRoute(start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.Route, error)
	SaveRoute(userID uuid.UUID, start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.SavedRoute, error)
	ListRoutes(userID uuid.UUID) ([]*entity.SavedRoute, error)
	ImportOSM(bbox string) error
	HandleBarrierEvent(event map[string]interface{}) error
}