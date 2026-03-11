// internal/service/geofence_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/andikatampubolon10/hris-backend/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GeofenceService interface {
	CreateGeofence(ctx context.Context, req models.CreateGeofenceRequest, createdBy string, createdByName string) (*models.GeofenceResponse, error)
	GetGeofenceByID(ctx context.Context, id string) (*models.GeofenceResponse, error)
	GetAllGeofences(ctx context.Context) ([]models.GeofenceResponse, error)
	GetActiveGeofences(ctx context.Context) ([]models.GeofenceResponse, error)
	UpdateGeofence(ctx context.Context, id string, req models.UpdateGeofenceRequest) (*models.GeofenceResponse, error)
	DeleteGeofence(ctx context.Context, id string) error
	CheckLocation(ctx context.Context, req models.CheckLocationRequest) (*models.CheckLocationResponse, error)
}

type geofenceService struct {
	geofenceRepo repository.GeofenceRepository
	userRepo     repository.UserRepository
}

func NewGeofenceService(
	geofenceRepo repository.GeofenceRepository,
	userRepo repository.UserRepository,
) GeofenceService {
	return &geofenceService{
		geofenceRepo: geofenceRepo,
		userRepo:     userRepo,
	}
}

func (s *geofenceService) CreateGeofence(ctx context.Context, req models.CreateGeofenceRequest, createdBy string, createdByName string) (*models.GeofenceResponse, error) {
	// Validate coordinates
	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, errors.New("latitude must be between -90 and 90")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, errors.New("longitude must be between -180 and 180")
	}

	// Convert creator ID
	creatorID, err := primitive.ObjectIDFromHex(createdBy)
	if err != nil {
		return nil, errors.New("invalid creator ID")
	}

	// Create GeoJSON Point
	geoPoint := models.GeoPoint{
		Type:        "Point",
		Coordinates: []float64{req.Longitude, req.Latitude}, // GeoJSON: [lng, lat]
	}

	// Convert department IDs
	var departmentIDs []primitive.ObjectID
	if len(req.DepartmentIDs) > 0 {
		departmentIDs = make([]primitive.ObjectID, len(req.DepartmentIDs))
		for i, idStr := range req.DepartmentIDs {
			id, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				return nil, fmt.Errorf("invalid department ID: %s", idStr)
			}
			departmentIDs[i] = id
		}
	}

	// Convert position IDs
	var positionIDs []primitive.ObjectID
	if len(req.PositionIDs) > 0 {
		positionIDs = make([]primitive.ObjectID, len(req.PositionIDs))
		for i, idStr := range req.PositionIDs {
			id, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				return nil, fmt.Errorf("invalid position ID: %s", idStr)
			}
			positionIDs[i] = id
		}
	}

	// Set defaults
	if req.Icon == "" {
		req.Icon = "📍"
	}
	if req.Color == "" {
		req.Color = "#3B82F6" // Blue
	}

	// Create geofence
	geofence := &models.Geofence{
		Name:          req.Name,
		Description:   req.Description,
		Location:      geoPoint,
		Address:       req.Address,
		Radius:        req.Radius,
		RadiusUnit:    "meter",
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Icon:          req.Icon,
		Color:         req.Color,
		AppliesTo:     req.AppliesTo,
		DepartmentIDs: departmentIDs,
		PositionIDs:   positionIDs,
		IsActive:      true,
		CreatedBy:     creatorID,
		CreatedByName: createdByName,
	}

	err = s.geofenceRepo.Create(ctx, geofence)
	if err != nil {
		return nil, errors.New("failed to create geofence")
	}

	response := geofence.ToResponse()
	return &response, nil
}

func (s *geofenceService) GetGeofenceByID(ctx context.Context, id string) (*models.GeofenceResponse, error) {
	geofence, err := s.geofenceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := geofence.ToResponse()
	return &response, nil
}

func (s *geofenceService) GetAllGeofences(ctx context.Context) ([]models.GeofenceResponse, error) {
	geofences, err := s.geofenceRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]models.GeofenceResponse, len(geofences))
	for i, gf := range geofences {
		responses[i] = gf.ToResponse()
	}

	return responses, nil
}

func (s *geofenceService) GetActiveGeofences(ctx context.Context) ([]models.GeofenceResponse, error) {
	geofences, err := s.geofenceRepo.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]models.GeofenceResponse, len(geofences))
	for i, gf := range geofences {
		responses[i] = gf.ToResponse()
	}

	return responses, nil
}

func (s *geofenceService) UpdateGeofence(ctx context.Context, id string, req models.UpdateGeofenceRequest) (*models.GeofenceResponse, error) {
	// Check if geofence exists
	_, err := s.geofenceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate coordinates if provided
	if req.Latitude != 0 {
		if req.Latitude < -90 || req.Latitude > 90 {
			return nil, errors.New("latitude must be between -90 and 90")
		}
	}
	if req.Longitude != 0 {
		if req.Longitude < -180 || req.Longitude > 180 {
			return nil, errors.New("longitude must be between -180 and 180")
		}
	}

	// Update geofence
	err = s.geofenceRepo.Update(ctx, id, &req)
	if err != nil {
		return nil, errors.New("failed to update geofence")
	}

	// Get updated geofence
	updated, err := s.geofenceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := updated.ToResponse()
	return &response, nil
}

func (s *geofenceService) DeleteGeofence(ctx context.Context, id string) error {
	// Check if geofence exists
	_, err := s.geofenceRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.geofenceRepo.Delete(ctx, id)
	if err != nil {
		return errors.New("failed to delete geofence")
	}

	return nil
}

func (s *geofenceService) CheckLocation(ctx context.Context, req models.CheckLocationRequest) (*models.CheckLocationResponse, error) {
	// Get all active geofences
	geofences, err := s.geofenceRepo.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	// If user ID provided, filter geofences that apply to user
	var applicableGeofences []models.Geofence
	if req.UserID != "" {
		user, err := s.userRepo.FindByID(ctx, req.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}

		// ✅ FIX: Correct function call with parentheses
		for _, gf := range geofences {
			if s.geofenceAppliesToUser(user, &gf) { // ✅ Added parentheses and proper parameters
				applicableGeofences = append(applicableGeofences, gf)
			}
		}
	} else {
		// No user ID, check all geofences
		applicableGeofences = geofences
	}

	// Check if location is within any geofence
	var closestGeofence *models.Geofence
	closestDistance := math.MaxFloat64 // ✅ Fixed: Added math. prefix

	for i := range applicableGeofences {
		gf := &applicableGeofences[i]
		distance := utils.CalculateDistance(
			req.Latitude,
			req.Longitude,
			gf.Latitude,
			gf.Longitude,
		)

		// Check if within geofence
		if distance <= float64(gf.Radius) {
			if closestGeofence == nil || distance < closestDistance {
				closestGeofence = gf
				closestDistance = distance
			}
		} else {
			// Track closest geofence even if not within
			if distance < closestDistance {
				closestDistance = distance
			}
		}
	}

	// Build response
	if closestGeofence != nil {
		gfResp := closestGeofence.ToResponse()
		return &models.CheckLocationResponse{
			IsWithinGeofence: true,
			Geofence:         &gfResp,
			Distance:         closestDistance,
			Message:          fmt.Sprintf("Dalam radius %s (%.0fm dari pusat)", closestGeofence.Name, closestDistance),
		}, nil
	}

	// Not within any geofence
	return &models.CheckLocationResponse{
		IsWithinGeofence: false,
		Geofence:         nil,
		Distance:         closestDistance,
		Message:          fmt.Sprintf("Di luar radius lokasi yang diizinkan (%.0fm dari lokasi terdekat)", closestDistance),
	}, nil
}

// ✅ FIX: Correct function signature and implementation
func (s *geofenceService) geofenceAppliesToUser(user *models.User, geofence *models.Geofence) bool {
	switch geofence.AppliesTo {
	case "all":
		return true
	case "departments":
		for _, deptID := range geofence.DepartmentIDs {
			if deptID == user.DepartmentID {
				return true
			}
		}
		return false
	case "positions":
		for _, posID := range geofence.PositionIDs {
			if posID == user.PositionID {
				return true
			}
		}
		return false
	default:
		return false
	}
}