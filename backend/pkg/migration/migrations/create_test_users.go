// pkg/migration/migrations/003_create_test_users.go
package migrations

import (
	"context"
	"log"
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
	description := "Create test users for development with complete profile"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		userCollection := db.Collection("users")

		// ✅ DROP old NIK index if it exists
		_, err := userCollection.Indexes().DropOne(ctx, "nik_1")
		if err != nil {
			// Ignore error if index doesn't exist
			log.Printf("Note: Could not drop nik_1 index (might not exist): %v", err)
		}

		// ✅ DROP collection to ensure clean state
		err = userCollection.Drop(ctx)
		if err != nil {
			log.Printf("Note: Could not drop users collection: %v", err)
		}

		// ✅ Create indexes with new field name
		indexModels := []mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "payroll_number", Value: 1}},       // ✅ New field name
				Options: options.Index().SetUnique(true).SetSparse(true), // ✅ Sparse allows null
			},
			{
				Keys: bson.D{{Key: "role", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "department_id", Value: 1}},
			},
			{
				Keys: bson.D{{Key: "is_active", Value: 1}},
			},
		}
		_, err = userCollection.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			return err
		}

		// Get HR department
		hrDept := db.Collection("departments").FindOne(ctx, bson.M{"code": "HR"})
		var dept models.Department
		if err := hrDept.Decode(&dept); err != nil {
			return err
		}

		// Get HR Manager position
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

		// Parse birth date
		birthDate, _ := time.Parse("2006-01-02", "1990-05-15")

		// Create test users with complete profile
		testUsers := []interface{}{
			// Manager HR
			models.User{
				ID:               primitive.NewObjectID(),
				PayrollNumber:    "PAY001",
				Email:            "manager.hr@company.com",
				Password:         hashedPassword,
				FullName:         "Budi Santoso",
				BirthDate:        birthDate,
				Religion:         "Islam",
				LastEducation:    "S1",
				YearEnrolled:     "2020",
				EmploymentStatus: "Tetap",
				DepartmentID:     dept.ID,
				DepartmentName:   dept.Name,
				PositionID:       pos.ID,
				PositionName:     pos.Name,
				Phone:            "+6281234567890",
				Address:          "Jl. Sudirman No. 123, Jakarta Pusat",
				Role:             models.RoleManagerHR,
				IsActive:         true,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		}

		// Get IT department for additional test users
		itDept := db.Collection("departments").FindOne(ctx, bson.M{"code": "IT"})
		var itDepartment models.Department
		if err := itDept.Decode(&itDepartment); err == nil {
			// Get IT Manager position
			itManagerPos := db.Collection("positions").FindOne(ctx, bson.M{
				"department_id": itDepartment.ID,
				"level":         3,
			})
			var itPos models.Position
			if err := itManagerPos.Decode(&itPos); err == nil {
				// Manager IT
				testUsers = append(testUsers, models.User{
					ID:               primitive.NewObjectID(),
					PayrollNumber:    "PAY002",
					Email:            "manager.it@company.com",
					Password:         hashedPassword,
					FullName:         "Siti Aminah",
					BirthDate:        birthDate,
					Religion:         "Islam",
					LastEducation:    "S1",
					YearEnrolled:     "2019",
					EmploymentStatus: "Tetap",
					DepartmentID:     itDepartment.ID,
					DepartmentName:   itDepartment.Name,
					PositionID:       itPos.ID,
					PositionName:     itPos.Name,
					Phone:            "+6281234567891",
					Address:          "Jl. Gatot Subroto No. 45, Jakarta Selatan",
					Role:             models.RoleManagerDepartemen,
					IsActive:         true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				})
			}

			// Get IT Staff position
			itStaffPos := db.Collection("positions").FindOne(ctx, bson.M{
				"department_id": itDepartment.ID,
				"level":         1,
			})
			var itStaffPosition models.Position
			if err := itStaffPos.Decode(&itStaffPosition); err == nil {
				// Admin IT
				testUsers = append(testUsers, models.User{
					ID:               primitive.NewObjectID(),
					PayrollNumber:    "PAY003",
					Email:            "admin.it@company.com",
					Password:         hashedPassword,
					FullName:         "Ahmad Yani",
					BirthDate:        birthDate,
					Religion:         "Islam",
					LastEducation:    "D3",
					YearEnrolled:     "2021",
					EmploymentStatus: "Kontrak",
					DepartmentID:     itDepartment.ID,
					DepartmentName:   itDepartment.Name,
					PositionID:       itStaffPosition.ID,
					PositionName:     itStaffPosition.Name,
					Phone:            "+6281234567892",
					Address:          "Jl. Thamrin No. 78, Jakarta Pusat",
					Role:             models.RoleAdminDepartemen,
					IsActive:         true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				})

				// Staff IT
				testUsers = append(testUsers, models.User{
					ID:               primitive.NewObjectID(),
					PayrollNumber:    "PAY004",
					Email:            "staf.it@company.com",
					Password:         hashedPassword,
					FullName:         "Dewi Lestari",
					BirthDate:        birthDate,
					Religion:         "Kristen",
					LastEducation:    "SMA",
					YearEnrolled:     "2022",
					EmploymentStatus: "Magang",
					DepartmentID:     itDepartment.ID,
					DepartmentName:   itDepartment.Name,
					PositionID:       itStaffPosition.ID,
					PositionName:     itStaffPosition.Name,
					Phone:            "+6281234567893",
					Address:          "Jl. Kuningan No. 99, Jakarta Selatan",
					Role:             models.RoleStaf,
					IsActive:         true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				})
			}
		}

		// Get Finance department for accountant
		finDept := db.Collection("departments").FindOne(ctx, bson.M{"code": "FIN"})
		var finDepartment models.Department
		if err := finDept.Decode(&finDepartment); err == nil {
			finPos := db.Collection("positions").FindOne(ctx, bson.M{
				"department_id": finDepartment.ID,
			})
			var finPosition models.Position
			if err := finPos.Decode(&finPosition); err == nil {
				// Accountant
				testUsers = append(testUsers, models.User{
					ID:               primitive.NewObjectID(),
					PayrollNumber:    "PAY005",
					Email:            "accountant@company.com",
					Password:         hashedPassword,
					FullName:         "Rina Susanti",
					BirthDate:        birthDate,
					Religion:         "Buddha",
					LastEducation:    "S1",
					YearEnrolled:     "2018",
					EmploymentStatus: "Tetap",
					DepartmentID:     finDepartment.ID,
					DepartmentName:   finDepartment.Name,
					PositionID:       finPosition.ID,
					PositionName:     finPosition.Name,
					Phone:            "+6281234567894",
					Address:          "Jl. Rasuna Said No. 111, Jakarta Selatan",
					Role:             models.RoleAccountant,
					IsActive:         true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				})
			}
		}

		_, err = userCollection.InsertMany(ctx, testUsers)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("users").Drop(ctx)
	}

	return version, name, description, up, down
}
