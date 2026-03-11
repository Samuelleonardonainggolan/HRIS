// pkg/migration/migrations/005_create_geofences.go
package migrations

import (
	"context"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateGeofences() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 5
	name := "create_geofences"
	description := "Create geofences collection with indexes and seed sample data"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("geofences")

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "name", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "is_active", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "location", Value: "2dsphere"}}, // Geospatial index
			},
			{
				Keys: bson.D{{Key: "created_by", Value: 1}},
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return err
		}

		// Get Manager HR for created_by
		var hrManager models.User
		err = db.Collection("users").FindOne(ctx, bson.M{"role": "manager_hr"}).Decode(&hrManager)
		if err != nil {
			// If no HR manager, skip seeding
			return nil
		}

		// Seed sample geofences for hotel
		geofences := []interface{}{
			// Main Hotel Location
			models.Geofence{
				ID:          primitive.NewObjectID(),
				Name:        "Hotel Labosta",
				Description: "Lokasi utama Hotel Labosta - area utama untuk absensi karyawan",
				Location: models.GeoPoint{
					Type:        "Point",
					Coordinates: []float64{106.8456, -6.2088}, // [longitude, latitude] - Jakarta example
				},
				Address:       "Jl. MH Thamrin No. 1, Jakarta Pusat, DKI Jakarta 10310",
				Radius:        150,
				RadiusUnit:    "meter",
				Latitude:      -6.2088,
				Longitude:     106.8456,
				Icon:          "🏨",
				Color:         "#3B82F6",
				AppliesTo:     "all",
				IsActive:      true,
				CreatedBy:     hrManager.ID,
				CreatedByName: hrManager.FullName,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},

			// Front Office Area
			models.Geofence{
				ID:          primitive.NewObjectID(),
				Name:        "Front Office Area",
				Description: "Area khusus untuk staff front office",
				Location: models.GeoPoint{
					Type:        "Point",
					Coordinates: []float64{106.8458, -6.2086},
				},
				Address:       "Lobby Hotel Labosta",
				Radius:        50,
				RadiusUnit:    "meter",
				Latitude:      -6.2086,
				Longitude:     106.8458,
				Icon:          "🏢",
				Color:         "#10B981",
				AppliesTo:     "departments",
				IsActive:      true,
				CreatedBy:     hrManager.ID,
				CreatedByName: hrManager.FullName,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},

			// Restaurant Area
			models.Geofence{
				ID:          primitive.NewObjectID(),
				Name:        "Restaurant & Bar Area",
				Description: "Area F&B untuk staff restaurant dan bar",
				Location: models.GeoPoint{
					Type:        "Point",
					Coordinates: []float64{106.8454, -6.2090},
				},
				Address:       "Restaurant Lantai 1, Hotel Labosta",
				Radius:        80,
				RadiusUnit:    "meter",
				Latitude:      -6.2090,
				Longitude:     106.8454,
				Icon:          "🍽️",
				Color:         "#F59E0B",
				AppliesTo:     "departments",
				IsActive:      true,
				CreatedBy:     hrManager.ID,
				CreatedByName: hrManager.FullName,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},

			// Housekeeping Area
			models.Geofence{
				ID:          primitive.NewObjectID(),
				Name:        "Housekeeping Central",
				Description: "Area housekeeping untuk staff kebersihan",
				Location: models.GeoPoint{
					Type:        "Point",
					Coordinates: []float64{106.8452, -6.2092},
				},
				Address:       "Basement, Hotel Labosta",
				Radius:        60,
				RadiusUnit:    "meter",
				Latitude:      -6.2092,
				Longitude:     106.8452,
				Icon:          "🧹",
				Color:         "#8B5CF6",
				AppliesTo:     "departments",
				IsActive:      true,
				CreatedBy:     hrManager.ID,
				CreatedByName: hrManager.FullName,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},

			// Office Area (for admin/management)
			models.Geofence{
				ID:          primitive.NewObjectID(),
				Name:        "Office Management",
				Description: "Area kantor untuk staff administrasi dan management",
				Location: models.GeoPoint{
					Type:        "Point",
					Coordinates: []float64{106.8460, -6.2085},
				},
				Address:       "Lantai 2, Office Wing, Hotel Labosta",
				Radius:        100,
				RadiusUnit:    "meter",
				Latitude:      -6.2085,
				Longitude:     106.8460,
				Icon:          "💼",
				Color:         "#EF4444",
				AppliesTo:     "all",
				IsActive:      true,
				CreatedBy:     hrManager.ID,
				CreatedByName: hrManager.FullName,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		_, err = collection.InsertMany(ctx, geofences)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("geofences").Drop(ctx)
	}

	return version, name, description, up, down
}