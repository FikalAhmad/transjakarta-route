package validation

import (
	"fmt"

	"transjakarta/internal/domain"
	"transjakarta/internal/geospatial"
)

type ValidationReport struct {
	TotalRoutes int      `json:"total_routes"`
	TotalTrips  int      `json:"total_trips"`
	TotalStops  int      `json:"total_stops"`
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
}

func ValidateRoutePlan(plan *domain.RoutePlanResult) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}
	if len(plan.Segments) == 0 {
		return fmt.Errorf("route plan has 0 segments")
	}

	for i, seg := range plan.Segments {
		switch seg.Type {
		case domain.SegmentTypeWalk:
			if seg.From == nil || seg.To == nil {
				return fmt.Errorf("walk segment %d missing from or to stop", i)
			}
			if seg.DistanceMeters < 0 || seg.DistanceMeters > 50000 {
				return fmt.Errorf("walk segment %d unrealistic distance: %.1fm", i, seg.DistanceMeters)
			}
		case domain.SegmentTypeTransit:
			if seg.RouteID == "" {
				return fmt.Errorf("transit segment %d missing route_id", i)
			}
			if seg.BoardingStop == nil || seg.AlightingStop == nil {
				return fmt.Errorf("transit segment %d missing boarding or alighting stop", i)
			}
			if len(seg.IntermediateStops) < 2 {
				return fmt.Errorf("transit segment %d has fewer than 2 stops", i)
			}
		case domain.SegmentTypeTransfer:
			if seg.AtStop == nil {
				return fmt.Errorf("transfer segment %d missing at_stop", i)
			}
		default:
			return fmt.Errorf("unknown segment type %s at index %d", seg.Type, i)
		}
	}
	return nil
}

func ValidateCoordinates(lat, lon float64) error {
	if lat < -90.0 || lat > 90.0 {
		return fmt.Errorf("invalid latitude: %f", lat)
	}
	if lon < -180.0 || lon > 180.0 {
		return fmt.Errorf("invalid longitude: %f", lon)
	}
	// Verify within reasonable Jakarta / Greater area bounds
	d := geospatial.HaversineDistance(lat, lon, -6.2088, 106.8456)
	if d > 120000.0 {
		return fmt.Errorf("coordinates (%.4f, %.4f) outside Greater Jakarta operational area (> 120km)", lat, lon)
	}
	return nil
}
