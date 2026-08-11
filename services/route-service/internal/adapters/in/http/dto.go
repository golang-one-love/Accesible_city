package http

type BuildRouteRequest struct {
	Start struct {
		Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
		Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	} `json:"start" validate:"required"`
	Finish struct {
		Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
		Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	} `json:"finish" validate:"required"`
	MobilityProfile string `json:"mobility_profile" validate:"required,oneof=wheelchair stroller elderly default"`
}

type BuildRouteResponse struct {
	Nodes []NodeResponse `json:"nodes"`
	TotalDistance float64 `json:"total_distance"`
	MaxSeverity   int     `json:"max_severity"`
}

type NodeResponse struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ImportOSMRequest struct {
	BBox string `json:"bbox" validate:"required"`
}