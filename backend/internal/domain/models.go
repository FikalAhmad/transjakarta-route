package domain

import "time"

type ServiceType string

const (
	ServiceBRT        ServiceType = "BRT"
	ServiceNonBRT     ServiceType = "NON_BRT"
	ServiceMikrotrans ServiceType = "MIKROTRANS"
	ServiceAmari      ServiceType = "AMARI"
)

type Route struct {
	RouteID        string      `json:"route_id"`
	RouteShortName string      `json:"route_short_name"`
	RouteLongName  string      `json:"route_long_name"`
	RouteColor     string      `json:"route_color"`
	ServiceType    ServiceType `json:"service_type"`
}

type Trip struct {
	TripID      string `json:"trip_id"`
	RouteID     string `json:"route_id"`
	ServiceID   string `json:"service_id"`
	DirectionID int    `json:"direction_id"`
	ShapeID     string `json:"shape_id"`
}

type Stop struct {
	StopID         string  `json:"stop_id"`
	StopName       string  `json:"stop_name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	LocationType   int     `json:"location_type"`
	DistanceMeters float64 `json:"distance_meters,omitempty"`
}

type StopTime struct {
	TripID       string `json:"trip_id"`
	StopID       string `json:"stop_id"`
	StopSequence int    `json:"stop_sequence"`
}

type ShapePoint struct {
	ShapeID         string  `json:"shape_id"`
	ShapePtLat      float64 `json:"shape_pt_lat"`
	ShapePtLon      float64 `json:"shape_pt_lon"`
	ShapePtSequence int     `json:"shape_pt_sequence"`
}

type DatasetMetadata struct {
	DatasetVersion string    `json:"dataset_version"`
	SourceFilename string    `json:"source_filename"`
	RowCount       int       `json:"row_count"`
	ImportedAt     time.Time `json:"imported_at"`
}
