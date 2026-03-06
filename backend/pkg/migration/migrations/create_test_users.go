// pkg/migration/migrations/003_create_test_users.go
package migrations

import (
	"context"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/auth"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTestUsers() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 3
	name := "create_test_users"
	description := "Create test users for development"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		userCollection := db.Collection("users")

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "nik", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		}
		_, err := userCollection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return err
		}

		// Get HR department and position
		hrDept := db.Collection("departments").FindOne(ctx, bson.M{"code": "HR"})
		var dept models.Department
		if err := hrDept.Decode(&dept); err != nil {
			return err
		}

		hrPos := db.Collection("positions").FindOne(ctx, bson.M{
			"code":          "HR-01",
			"department_id": dept.ID,
		})
		var pos models.Position
		if err := hrPos.Decode(&pos); err != nil {
			return err
		}

		// Hash password
		hashedPassword, err := auth.HashPassword("password123")
		if err != nil {
			return err
		}

		// Create test user (menyesuaikan model User terbaru)
		testUser := models.User{
			ID:         primitive.NewObjectID(),
			NIK:        "EMP001",
			Email:      "manager.hr@company.com",
			Password:   hashedPassword,
			FullName:   "Manager HR",
			Role:       models.RoleManagerHR,
			Department: dept.Code,
			Position:   pos.Name,
			Phone:      "+6281234567890",
			JoinDate:   time.Now(),
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_, err = userCollection.InsertOne(ctx, testUser)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		_, err := db.Collection("users").DeleteMany(ctx, bson.M{"nik": "EMP001"})
		return err
	}

	return version, name, description, up, down
}
