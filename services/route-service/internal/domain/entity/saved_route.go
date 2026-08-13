package entity

import (
	"time"

	"github.com/google/uuid"
)

type SavedRoute struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	StartLat        float64
	StartLon        float64
	FinishLat       float64
	FinishLon       float64
	MobilityProfile MobilityProfile
	Points          []*Node
	TotalDistance   float64
	MaxSeverity     int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewSavedRoute(userID uuid.UUID, route *Route, profile MobilityProfile, startLat, startLon, finishLat, finishLon float64) *SavedRoute {
	now := time.Now().UTC()
	return &SavedRoute{
		ID:              uuid.New(),
		UserID:          userID,
		StartLat:        startLat,
		StartLon:        startLon,
		FinishLat:       finishLat,
		FinishLon:       finishLon,
		MobilityProfile: profile,
		Points:          route.Nodes,
		TotalDistance:   route.TotalDistance,
		MaxSeverity:     route.MaxSeverity,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (r *SavedRoute) Update(route *Route) {
	r.Points = route.Nodes
	r.TotalDistance = route.TotalDistance
	r.MaxSeverity = route.MaxSeverity
	r.UpdatedAt = time.Now().UTC()
}