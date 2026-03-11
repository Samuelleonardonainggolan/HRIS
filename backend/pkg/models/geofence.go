// pkg/models/geofence.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Geofence represents a geofence location for attendance
type Geofence struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"` // Nama lokasi (e.g., "Hotel Labosta")
	Description string             `json:"description" bson:"description"`
	
	// Location
	Location GeoPoint `json:"location" bson:"location"` // Center point
	Address  string   `json:"address" bson:"address"`   // Full address
	
	// Geofence Settings
	Radius      int    `json:"radius" bson:"radius"`           // Radius in meters
	RadiusUnit  string `json:"radius_unit" bson:"radius_unit"` // "meter" or "km"
	
	// Coordinates
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	
	// Icon & Color for UI
	Icon  string `json:"icon" bson:"icon"`   // Emoji or icon identifier
	Color string `json:"color" bson:"color"` // Hex color for map marker
	
	// Assignment
	AppliesTo      string               `json:"applies_to" bson:"applies_to"` // "all", "departments", "positions"
	DepartmentIDs  []primitive.ObjectID `json:"department_ids,omitempty" bson:"department_ids,omitempty"`
	PositionIDs    []primitive.ObjectID `json:"position_ids,omitempty" bson:"position_ids,omitempty"`
	
	// Status
	IsActive bool `json:"is_active" bson:"is_active"`
	
	// Metadata
	CreatedBy   primitive.ObjectID `json:"created_by" bson:"created_by"`
	CreatedByName string           `json:"created_by_name" bson:"created_by_name"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// GeoPoint represents GeoJSON Point for MongoDB geospatial queries
type GeoPoint struct {
	Type        string    `json:"type" bson:"type"`               // Always "Point"
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // [longitude, latitude]
}

// CreateGeofenceRequest represents request to create geofence
type CreateGeofenceRequest struct {
	Name          string               `json:"name" binding:"required"`
	Description   string               `json:"description"`
	Latitude      float64              `json:"latitude" binding:"required"`
	Longitude     float64              `json:"longitude" binding:"required"`
	Address       string               `json:"address"`
	Radius        int                  `json:"radius" binding:"required,min=10,max=5000"` // 10m - 5km
	Icon          string               `json:"icon"`
	Color         string               `json:"color"`
	AppliesTo     string               `json:"applies_to" binding:"required,oneof=all departments positions"`
	DepartmentIDs []string             `json:"department_ids"`
	PositionIDs   []string             `json:"position_ids"`
}

// UpdateGeofenceRequest represents request to update geofence
type UpdateGeofenceRequest struct {
	Name          string   `json:"name,omitempty"`
	Description   string   `json:"description,omitempty"`
	Latitude      float64  `json:"latitude,omitempty"`
	Longitude     float64  `json:"longitude,omitempty"`
	Address       string   `json:"address,omitempty"`
	Radius        int      `json:"radius,omitempty"`
	Icon          string   `json:"icon,omitempty"`
	Color         string   `json:"color,omitempty"`
	AppliesTo     string   `json:"applies_to,omitempty"`
	DepartmentIDs []string `json:"department_ids,omitempty"`
	PositionIDs   []string `json:"position_ids,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

// GeofenceResponse represents geofence response
type GeofenceResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Address       string    `json:"address"`
	Radius        int       `json:"radius"`
	RadiusUnit    string    `json:"radius_unit"`
	Icon          string    `json:"icon"`
	Color         string    `json:"color"`
	AppliesTo     string    `json:"applies_to"`
	DepartmentIDs []string  `json:"department_ids,omitempty"`
	PositionIDs   []string  `json:"position_ids,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedBy     string    `json:"created_by"`
	CreatedByName string    `json:"created_by_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CheckLocationRequest represents request to check if location is within geofence
type CheckLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	UserID    string  `json:"user_id,omitempty"`
}

// CheckLocationResponse represents response for location check
type CheckLocationResponse struct {
	IsWithinGeofence bool               `json:"is_within_geofence"`
	Geofence         *GeofenceResponse  `json:"geofence,omitempty"`
	Distance         float64            `json:"distance"` // Distance in meters
	Message          string             `json:"message"`
}

// ToResponse converts Geofence to GeofenceResponse
func (g *Geofence) ToResponse() GeofenceResponse {
	resp := GeofenceResponse{
		ID:            g.ID.Hex(),
		Name:          g.Name,
		Description:   g.Description,
		Latitude:      g.Latitude,
		Longitude:     g.Longitude,
		Address:       g.Address,
		Radius:        g.Radius,
		RadiusUnit:    g.RadiusUnit,
		Icon:          g.Icon,
		Color:         g.Color,
		AppliesTo:     g.AppliesTo,
		IsActive:      g.IsActive,
		CreatedBy:     g.CreatedBy.Hex(),
		CreatedByName: g.CreatedByName,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}

	// Convert ObjectIDs to strings
	if len(g.DepartmentIDs) > 0 {
		resp.DepartmentIDs = make([]string, len(g.DepartmentIDs))
		for i, id := range g.DepartmentIDs {
			resp.DepartmentIDs[i] = id.Hex()
		}
	}

	if len(g.PositionIDs) > 0 {
		resp.PositionIDs = make([]string, len(g.PositionIDs))
		for i, id := range g.PositionIDs {
			resp.PositionIDs[i] = id.Hex()
		}
	}

	return resp
}