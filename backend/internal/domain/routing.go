package domain

type SegmentType string

const (
	SegmentTypeWalk     SegmentType = "WALK"
	SegmentTypeTransit  SegmentType = "TRANSIT"
	SegmentTypeTransfer SegmentType = "TRANSFER"
)

type Preference string

const (
	PreferenceBalanced       Preference = "BALANCED"
	PreferenceFastest        Preference = "FASTEST"
	PreferenceFewestTransfers Preference = "FEWEST_TRANSFERS"
	PreferenceLeastWalking   Preference = "LEAST_WALKING"
)

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RouteSearchRequest struct {
	Origin      Coordinate `json:"origin"`
	Destination Coordinate `json:"destination"`
	Preference  Preference `json:"preference"`
}

type GeoJSONGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][]float64   `json:"coordinates"`
}

type Segment struct {
	Type                 SegmentType      `json:"type"`
	Description          string           `json:"description"`
	DistanceMeters       float64          `json:"distance_meters"`
	DurationMinutes      float64          `json:"duration_minutes"`
	From                 *Stop            `json:"from,omitempty"`
	To                   *Stop            `json:"to,omitempty"`
	AtStop               *Stop            `json:"at_stop,omitempty"`
	RouteID              string           `json:"route_id,omitempty"`
	RouteName            string           `json:"route_name,omitempty"`
	RouteColor           string           `json:"route_color,omitempty"`
	ServiceType          ServiceType      `json:"service_type,omitempty"`
	DirectionID          int              `json:"direction_id,omitempty"`
	BoardingStop         *Stop            `json:"boarding_stop,omitempty"`
	AlightingStop        *Stop            `json:"alighting_stop,omitempty"`
	IntermediateStops    []Stop           `json:"intermediate_stops,omitempty"`
	Geometry             *GeoJSONGeometry `json:"geometry,omitempty"`
}

type RoutePlanResult struct {
	RouteID               string           `json:"route_id"`
	Summary               string           `json:"summary"`
	TotalDurationMinutes  float64          `json:"total_duration_minutes"`
	TotalDistanceMeters   float64          `json:"total_distance_meters"`
	WalkingDistanceMeters float64          `json:"walking_distance_meters"`
	TransitDistanceMeters float64          `json:"transit_distance_meters"`
	TransferCount         int              `json:"transfer_count"`
	TotalFare             int              `json:"total_fare"`
	Segments              []Segment        `json:"segments"`
	FullGeometry          *GeoJSONGeometry `json:"full_geometry,omitempty"`
}

type RouteSearchResponse struct {
	Data struct {
		Routes []RoutePlanResult `json:"routes"`
	} `json:"data"`
}
