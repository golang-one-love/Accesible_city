package http

import (
	"net/http"
	"strconv"

	"github.com/accessible-path/route-service/internal/domain/entity"
	"github.com/accessible-path/route-service/internal/domain/valueobject"
	"github.com/accessible-path/route-service/internal/ports/in"
	"github.com/google/uuid"
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

	protected := g.Group("", JWTUserIDMiddleware())
	protected.POST("/save", h.SaveRoute)
	protected.GET("", h.ListRoutes)
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

func (h *RouteHandler) SaveRoute(c echo.Context) error {
	userID, ok := userIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req SaveRouteRequest
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

	saved, err := h.routeUseCase.SaveRoute(userID, start, finish, profile)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, toSavedRouteResponse(saved))
}

func (h *RouteHandler) ListRoutes(c echo.Context) error {
	userID, ok := userIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	routes, err := h.routeUseCase.ListRoutes(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	resp := make([]SavedRouteResponse, len(routes))
	for i, r := range routes {
		resp[i] = toSavedRouteResponse(r)
	}

	return c.JSON(http.StatusOK, SavedRoutesResponse{Routes: resp, Total: len(resp)})
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
			ID:        "osm:" + strconv.FormatInt(n.ID, 10),
			Latitude:  n.Latitude,
			Longitude: n.Longitude,
		}
	}
	return resp
}

func toSavedRouteResponse(r *entity.SavedRoute) SavedRouteResponse {
	return SavedRouteResponse{
		ID:              r.ID.String(),
		Start:           CoordinatesDTO{Latitude: r.StartLat, Longitude: r.StartLon},
		Finish:          CoordinatesDTO{Latitude: r.FinishLat, Longitude: r.FinishLon},
		MobilityProfile: string(r.MobilityProfile),
		Points:          toNodeResponses(r.Points),
		TotalDistance:   r.TotalDistance,
		MaxSeverity:     r.MaxSeverity,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func userIDFromContext(c echo.Context) (uuid.UUID, bool) {
	id, ok := c.Get("user_id").(string)
	if !ok || id == "" {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, false
	}
	return parsed, true
}
