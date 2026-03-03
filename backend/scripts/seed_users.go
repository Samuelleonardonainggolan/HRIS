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
            NIK:        "EMP001",
            Email:      "manager.hr@company.com",
            Password:   hashedPassword,
            FullName:   "Manager HR",
            Role:       models.RoleManagerHR,
            Department: "HR",
            Position:   "HR Manager",
            Phone:      "+6281234567890",
            JoinDate:   time.Now(),
            IsActive:   true,
        },
        {
            NIK:        "EMP002",
            Email:      "manager.it@company.com",
            Password:   hashedPassword,
            FullName:   "Manager IT",
            Role:       models.RoleManagerDept,
            Department: "IT",
            Position:   "IT Manager",
            Phone:      "+6281234567891",
            JoinDate:   time.Now(),
            IsActive:   true,
        },
        {
            NIK:        "EMP003",
            Email:      "admin.it@company.com",
            Password:   hashedPassword,
            FullName:   "Admin IT",
            Role:       models.RoleAdminDept,
            Department: "IT",
            Position:   "IT Admin",
            Phone:      "+6281234567892",
            JoinDate:   time.Now(),
            IsActive:   true,
        },
        {
            NIK:        "EMP004",
            Email:      "staf.it@company.com",
            Password:   hashedPassword,
            FullName:   "Staf IT",
            Role:       models.RoleStaf,
            Department: "IT",
            Position:   "Junior Developer",
            Phone:      "+6281234567893",
            JoinDate:   time.Now(),
            IsActive:   true,
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
    log.Println("  Email: manager.hr@company.com | Password: password123 | Role: Manager HR")
    log.Println("  Email: manager.it@company.com | Password: password123 | Role: Manager Departemen")
    log.Println("  Email: admin.it@company.com   | Password: password123 | Role: Admin Departemen")
    log.Println("  Email: staf.it@company.com    | Password: password123 | Role: Staf")
}