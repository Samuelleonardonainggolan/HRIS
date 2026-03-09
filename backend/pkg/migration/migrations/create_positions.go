// pkg/migration/migrations/002_create_positions.go
package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreatePositions() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 2
	name := "create_positions"
	description := "Create positions collection and seed initial positions"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("positions")

		// Create indexes
		indexModels := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "code", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "department_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "level", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "is_active", Value: 1}},
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Get HR department
		var hrDepartment models.Department
		err = db.Collection("departments").FindOne(ctx, bson.M{"code": "HR"}).Decode(&hrDepartment)
		if err != nil {
			return fmt.Errorf("HR department not found (code: HR). Make sure migration 1 ran successfully: %w", err)
		}

		// Get IT department
		var itDepartment models.Department
		err = db.Collection("departments").FindOne(ctx, bson.M{"code": "IT"}).Decode(&itDepartment)
		if err != nil {
			return fmt.Errorf("IT department not found (code: IT). Make sure migration 1 ran successfully: %w", err)
		}

		// Get Finance department
		var finDepartment models.Department
		err = db.Collection("departments").FindOne(ctx, bson.M{"code": "FIN"}).Decode(&finDepartment)
		if err != nil {
			return fmt.Errorf("Finance department not found (code: FIN). Make sure migration 1 ran successfully: %w", err)
		}

		// Seed positions
		positions := []interface{}{
			// HR Positions
			models.Position{
				ID:           primitive.NewObjectID(),
				Code:         "HR-01",
				Name:         "HR Manager",
				DepartmentID: hrDepartment.ID,
				Level:        models.LevelManager,
				Description:  "Manages human resources department",
				Requirements: "Bachelor degree, 5+ years experience in HR",
				SalaryRange: models.SalaryRange{
					Min:      10000000,
					Max:      15000000,
					Currency: "IDR",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			models.Position{
				ID:           primitive.NewObjectID(),
				Code:         "HR-02",
				Name:         "HR Staff",
				DepartmentID: hrDepartment.ID,
				Level:        models.LevelStaff,
				Description:  "Handles recruitment and employee administration",
				Requirements: "Bachelor degree, 1+ years experience",
				SalaryRange: models.SalaryRange{
					Min:      5000000,
					Max:      7000000,
					Currency: "IDR",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			// IT Positions
			models.Position{
				ID:           primitive.NewObjectID(),
				Code:         "IT-01",
				Name:         "IT Manager",
				DepartmentID: itDepartment.ID,
				Level:        models.LevelManager,
				Description:  "Manages IT infrastructure and team",
				Requirements: "Bachelor degree in IT, 5+ years experience",
				SalaryRange: models.SalaryRange{
					Min:      12000000,
					Max:      18000000,
					Currency: "IDR",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			models.Position{
				ID:           primitive.NewObjectID(),
				Code:         "IT-02",
				Name:         "Software Developer",
				DepartmentID: itDepartment.ID,
				Level:        models.LevelStaff,
				Description:  "Develops and maintains software applications",
				Requirements: "Bachelor degree in IT, 2+ years experience",
				SalaryRange: models.SalaryRange{
					Min:      7000000,
					Max:      10000000,
					Currency: "IDR",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			models.Position{
				ID:           primitive.NewObjectID(),
				Code:         "IT-03",
				Name:         "System Administrator",
				DepartmentID: itDepartment.ID,
				Level:        models.LevelStaff,
				Description:  "Maintains servers and network infrastructure",
				Requirements: "Bachelor degree in IT, 1+ years experience",
				SalaryRange: models.SalaryRange{
					Min:      6000000,
					Max:      9000000,
					Currency: "IDR",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			// Finance Positions
			models.Position{
				ID:           primitive.NewObjectID(),
				Code:         "FIN-01",
				Name:         "Accountant",
				DepartmentID: finDepartment.ID,
				Level:        models.LevelStaff,
				Description:  "Handles accounting and financial reporting",
				Requirements: "Bachelor degree in Accounting, 2+ years experience",
				SalaryRange: models.SalaryRange{
					Min:      6000000,
					Max:      9000000,
					Currency: "IDR",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		_, err = collection.InsertMany(ctx, positions)
		if err != nil {
			return fmt.Errorf("failed to insert positions: %w", err)
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("positions").Drop(ctx)
	}

	return version, name, description, up, down
}