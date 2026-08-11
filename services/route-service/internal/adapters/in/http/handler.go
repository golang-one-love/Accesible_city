package http

import (
	"net/http"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/in"
	"github.com/labstack/echo/v4"
)

type RouteHandler struct {
	routeUseCase in.RouteUseCase
}

func NewRouteHandler(routeUseCase in.RouteUseCase) *RouteHandler {
	return &RouteHandler{routeUseCase: routeUseCase}
}

func (h *RouteHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/build", h.BuildRoute)
	g.POST("/admin/import-osm", h.ImportOSM)
}

func (h *RouteHandler) BuildRoute(c echo.Context) error {
	var req BuildRouteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	start, err := valueobject.NewCoordinates(req.Start.Latitude, req.Start.Longitude)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start coordinates"})
	}

	finish, err := valueobject.NewCoordinates(req.Finish.Latitude, req.Finish.Longitude)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid finish coordinates"})
	}

	profile := entity.MobilityProfile(req.MobilityProfile)

	route, err := h.routeUseCase.BuildRoute(start, finish, profile)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, BuildRouteResponse{
		Nodes:         toNodeResponses(route.Nodes),
		TotalDistance: route.TotalDistance,
		MaxSeverity:   route.MaxSeverity,
	})
}

func (h *RouteHandler) ImportOSM(c echo.Context) error {
	var req ImportOSMRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if err := h.routeUseCase.ImportOSM(req.BBox); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "import started"})
}

func toNodeResponses(nodes []*entity.Node) []NodeResponse {
	resp := make([]NodeResponse, len(nodes))
	for i, n := range nodes {
		resp[i] = NodeResponse{
			ID:        n.ID,
			Latitude:  n.Latitude,
			Longitude: n.Longitude,
		}
	}
	return resp
}