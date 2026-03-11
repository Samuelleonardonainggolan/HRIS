//go:build seed_users

// scripts/seed_users.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/config"
	"github.com/andikatampubolon10/hris-backend/pkg/auth"
	"github.com/andikatampubolon10/hris-backend/pkg/database"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	cfg := config.LoadConfig()

	mongodb, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer mongodb.Disconnect()

	userRepo := repository.NewUserRepository(mongodb.Database)
	ctx := context.Background()

	// Hash password
	hashedPassword, _ := auth.HashPassword("password123")

	users := []*models.User{
		{
			ID:               primitive.NewObjectID(),
			PayrollNumber:    "EMP001",
			Email:            "manager.hr@company.com",
			Password:         hashedPassword,
			FullName:         "Manager HR",
			BirthDate:        time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
			Religion:         "Islam",
			LastEducation:    "S1 Psikologi",
			YearEnrolled:     "2010",
			EmploymentStatus: "Permanent",
			DepartmentName:   "HR",
			PositionName:     "HR Manager",
			Phone:            "+6281234567890",
			Address:          "Jl. HR Rasuna Said, Jakarta",
			Role:             models.RoleManagerHR,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               primitive.NewObjectID(),
			PayrollNumber:    "EMP002",
			Email:            "manager.it@company.com",
			Password:         hashedPassword,
			FullName:         "Manager IT",
			BirthDate:        time.Date(1988, 8, 20, 0, 0, 0, 0, time.UTC),
			Religion:         "Kristen",
			LastEducation:    "S1 Teknik Informatika",
			YearEnrolled:     "2012",
			EmploymentStatus: "Permanent",
			DepartmentName:   "IT",
			PositionName:     "IT Manager",
			Phone:            "+6281234567891",
			Address:          "Jl. Gatot Subroto, Jakarta",
			Role:             models.RoleManagerDepartemen,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               primitive.NewObjectID(),
			PayrollNumber:    "EMP003",
			Email:            "admin.it@company.com",
			Password:         hashedPassword,
			FullName:         "Admin IT",
			BirthDate:        time.Date(1992, 3, 10, 0, 0, 0, 0, time.UTC),
			Religion:         "Islam",
			LastEducation:    "S1 Sistem Informasi",
			YearEnrolled:     "2015",
			EmploymentStatus: "Permanent",
			DepartmentName:   "IT",
			PositionName:     "IT Admin",
			Phone:            "+6281234567892",
			Address:          "Jl. Sudirman, Jakarta",
			Role:             models.RoleAdminDepartemen,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               primitive.NewObjectID(),
			PayrollNumber:    "EMP004",
			Email:            "staf.it@company.com",
			Password:         hashedPassword,
			FullName:         "Staf IT",
			BirthDate:        time.Date(1995, 11, 25, 0, 0, 0, 0, time.UTC),
			Religion:         "Buddha",
			LastEducation:    "S1 Teknik Informatika",
			YearEnrolled:     "2018",
			EmploymentStatus: "Contract",
			DepartmentName:   "IT",
			PositionName:     "Junior Developer",
			Phone:            "+6281234567893",
			Address:          "Jl. Thamrin, Jakarta",
			Role:             models.RoleStaf,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	for _, user := range users {
		err := userRepo.Create(ctx, user)
		if err != nil {
			log.Printf("❌ Failed to create %s: %v", user.Email, err)
		} else {
			log.Printf("✅ Created user: %s (%s)", user.Email, user.Role)
		}
	}

	log.Println("\n🎉 Seed completed!")
	log.Println("\n📋 Test Users:")
	log.Println("  Email: manager.hr@company.com | Password: password123 | Role: manager_hr")
	log.Println("  Email: manager.it@company.com | Password: password123 | Role: manager_departemen")
	log.Println("  Email: admin.it@company.com   | Password: password123 | Role: admin_departemen")
	log.Println("  Email: staf.it@company.com    | Password: password123 | Role: staf")
}
