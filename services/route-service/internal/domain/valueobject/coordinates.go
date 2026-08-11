package valueobject

import (
	"errors"
	"math"
)

var ErrInvalidCoordinates = errors.New("invalid coordinates")

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func NewCoordinates(lat, lon float64) (Coordinates, error) {
	if lat < -90 || lat > 90 {
		return Coordinates{}, ErrInvalidCoordinates
	}
	if lon < -180 || lon > 180 {
		return Coordinates{}, ErrInvalidCoordinates
	}
	return Coordinates{Latitude: lat, Longitude: lon}, nil
}

func (c Coordinates) HaversineDistance(other Coordinates) float64 {
	const R = 6371000 // Earth radius in meters
	lat1 := c.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	dLat := (other.Latitude - c.Latitude) * math.Pi / 180
	dLon := (other.Longitude - c.Longitude) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}