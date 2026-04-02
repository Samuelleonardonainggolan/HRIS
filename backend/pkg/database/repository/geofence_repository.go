// pkg/database/repository/geofence_repository.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GeofenceRepository interface {
	Create(ctx context.Context, geofence *models.Geofence) error
	FindByID(ctx context.Context, id string) (*models.Geofence, error)
	FindAll(ctx context.Context) ([]models.Geofence, error)
	FindActive(ctx context.Context) ([]models.Geofence, error)
	FindByLocation(ctx context.Context, latitude, longitude float64, maxDistance int) ([]models.Geofence, error)
	Update(ctx context.Context, id string, req *models.UpdateGeofenceRequest) error
	Delete(ctx context.Context, id string) error
	CheckUserInGeofence(ctx context.Context, userID string, latitude, longitude float64) (*models.Geofence, float64, error)
}

type geofenceRepository struct {
	collection *mongo.Collection
}

func NewGeofenceRepository(db *mongo.Database) GeofenceRepository {
	return &geofenceRepository{
		collection: db.Collection("geofences"),
	}
}

func (r *geofenceRepository) Create(ctx context.Context, geofence *models.Geofence) error {
	geofence.ID = primitive.NewObjectID()
	geofence.CreatedAt = time.Now()
	geofence.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, geofence)
	return err
}

func (r *geofenceRepository) FindByID(ctx context.Context, id string) (*models.Geofence, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid geofence ID")
	}

	var geofence models.Geofence
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&geofence)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("geofence not found")
		}
		return nil, err
	}

	return &geofence, nil
}

func (r *geofenceRepository) FindAll(ctx context.Context) ([]models.Geofence, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var geofences []models.Geofence
	if err = cursor.All(ctx, &geofences); err != nil {
		return nil, err
	}

	return geofences, nil
}

func (r *geofenceRepository) FindActive(ctx context.Context) ([]models.Geofence, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var geofences []models.Geofence
	if err = cursor.All(ctx, &geofences); err != nil {
		return nil, err
	}

	return geofences, nil
}

func (r *geofenceRepository) FindByLocation(ctx context.Context, latitude, longitude float64, maxDistance int) ([]models.Geofence, error) {
	// Use MongoDB $geoNear or $near query
	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{longitude, latitude}, // GeoJSON uses [lng, lat]
				},
				"$maxDistance": maxDistance, // in meters
			},
		},
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var geofences []models.Geofence
	if err = cursor.All(ctx, &geofences); err != nil {
		return nil, err
	}

	return geofences, nil
}

func (r *geofenceRepository) Update(ctx context.Context, id string, req *models.UpdateGeofenceRequest) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid geofence ID")
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Build update document
	setFields := update["$set"].(bson.M)

	if req.Name != "" {
		setFields["name"] = req.Name
	}
	if req.Description != "" {
		setFields["description"] = req.Description
	}
	if req.Latitude != 0 {
		setFields["latitude"] = req.Latitude
		setFields["location.coordinates.1"] = req.Latitude
	}
	if req.Longitude != 0 {
		setFields["longitude"] = req.Longitude
		setFields["location.coordinates.0"] = req.Longitude
	}
	if req.Address != "" {
		setFields["address"] = req.Address
	}
	if req.Radius > 0 {
		setFields["radius"] = req.Radius
	}
	if req.Icon != "" {
		setFields["icon"] = req.Icon
	}
	if req.Color != "" {
		setFields["color"] = req.Color
	}
	if req.AppliesTo != "" {
		setFields["applies_to"] = req.AppliesTo
	}
	if req.IsActive != nil {
		setFields["is_active"] = *req.IsActive
	}

	// Handle arrays
	if len(req.DepartmentIDs) > 0 {
		ids := make([]primitive.ObjectID, len(req.DepartmentIDs))
		for i, idStr := range req.DepartmentIDs {
			id, _ := primitive.ObjectIDFromHex(idStr)
			ids[i] = id
		}
		setFields["department_ids"] = ids
	}

	if len(req.PositionIDs) > 0 {
		ids := make([]primitive.ObjectID, len(req.PositionIDs))
		for i, idStr := range req.PositionIDs {
			id, _ := primitive.ObjectIDFromHex(idStr)
			ids[i] = id
		}
		setFields["position_ids"] = ids
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("geofence not found")
	}

	return nil
}

func (r *geofenceRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid geofence ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("geofence not found")
	}

	return nil
}

func (r *geofenceRepository) CheckUserInGeofence(ctx context.Context, userID string, latitude, longitude float64) (*models.Geofence, float64, error) {
	// This would check if user's location is within any geofence that applies to them
	// For now, return all active geofences and check distance in service layer
	geofences, err := r.FindActive(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Find the closest geofence that user is within
	for _, gf := range geofences {
		distance := calculateDistance(latitude, longitude, gf.Latitude, gf.Longitude)
		if distance <= float64(gf.Radius) {
			return &gf, distance, nil
		}
	}

	return nil, 0, nil
}

// calculateDistance calculates distance between two coordinates using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Implementation in next file
	return 0
}
