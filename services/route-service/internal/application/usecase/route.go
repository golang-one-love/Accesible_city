package usecase

import (
	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/service"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/in"
)

type routeUseCase struct {
	router *service.Router
}

func NewRouteUseCase(router *service.Router) in.RouteUseCase {
	return &routeUseCase{router: router}
}

func (u *routeUseCase) BuildRoute(start, finish valueobject.Coordinates, profile entity.MobilityProfile) (*entity.Route, error) {
	return u.router.BuildRoute(start, finish, profile)
}

func (u *routeUseCase) ImportOSM(bbox string) error {
	return nil // TODO: implement OSM import
}