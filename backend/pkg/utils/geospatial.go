// pkg/utils/geospatial.go
package utils

import (
	"fmt"
	"math"
)

const (
	// EarthRadiusKm is Earth's radius in kilometers
	EarthRadiusKm = 6371.0
	// EarthRadiusMeters is Earth's radius in meters
	EarthRadiusMeters = 6371000.0
)

// CalculateDistance calculates distance between two coordinates using Haversine formula
// Returns distance in meters
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in meters
	distance := EarthRadiusMeters * c

	return distance
}

// IsWithinRadius checks if a point is within radius of center point
func IsWithinRadius(centerLat, centerLon, pointLat, pointLon float64, radiusMeters int) bool {
	distance := CalculateDistance(centerLat, centerLon, pointLat, pointLon)
	return distance <= float64(radiusMeters)
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// radiansToDegrees converts radians to degrees
func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// GetBoundingBox returns bounding box coordinates for a given center and radius
// Returns (minLat, minLon, maxLat, maxLon)
func GetBoundingBox(centerLat, centerLon float64, radiusMeters int) (float64, float64, float64, float64) {
	// Angular distance in radians on a great circle
	radDist := float64(radiusMeters) / EarthRadiusMeters

	minLat := centerLat - radiansToDegrees(radDist)
	maxLat := centerLat + radiansToDegrees(radDist)

	// Compensate for degrees longitude getting smaller with latitude
	minLon := centerLon - radiansToDegrees(radDist/math.Cos(degreesToRadians(centerLat)))
	maxLon := centerLon + radiansToDegrees(radDist/math.Cos(degreesToRadians(centerLat)))

	return minLat, minLon, maxLat, maxLon
}

// FormatDistance formats distance in meters to human-readable string
func FormatDistance(meters float64) string {
	if meters < 1000 {
		// ✅ FIX: Convert float64 to string using fmt.Sprintf
		return fmt.Sprintf("%.0fm", math.Round(meters))
	}
	km := meters / 1000
	// ✅ FIX: Format with 2 decimal places
	return fmt.Sprintf("%.2fkm", math.Round(km*100)/100)
}