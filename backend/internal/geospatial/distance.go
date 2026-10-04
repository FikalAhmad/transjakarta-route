package geospatial

import (
	"math"
)

const EarthRadiusMeters = 6371000.0

// HaversineDistance calculates the geodesic distance in meters between two lat/lon pairs
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusMeters * c
}

// EstimateWalkMinutes calculates walking duration assuming 4.5 km/h (75 m/min)
func EstimateWalkMinutes(distanceMeters float64) float64 {
	return distanceMeters / 75.0
}

// EstimateTransitMinutes calculates transit duration assuming average transit speed 20 km/h (333 m/min) + dwell
func EstimateTransitMinutes(distanceMeters float64, stopCount int) float64 {
	travelMin := distanceMeters / 333.33
	dwellMin := float64(stopCount) * 0.75 // 45 seconds per stop dwell time
	return travelMin + dwellMin
}
