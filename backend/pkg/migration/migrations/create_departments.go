// pkg/migration/migrations/001_create_departments.go
package migrations

import (
    "context"
    "time"

    "github.com/andikatampubolon10/hris-backend/pkg/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func CreateDepartments() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
    version := 1
    name := "create_departments"
    description := "Create departments collection and seed initial data"

    up := func(db *mongo.Database) error {
        ctx := context.Background()
        collection := db.Collection("departments")

        // Create unique index on code
        indexModel := mongo.IndexModel{
            Keys:    bson.D{{Key: "code", Value: 1}},
            Options: options.Index().SetUnique(true),
        }
        _, err := collection.Indexes().CreateOne(ctx, indexModel)
        if err != nil {
            return err
        }

        // Seed departments
        departments := []interface{}{
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "FO",
                Name:        "Front Office",
                Description: "Layanan penerimaan tamu, reservasi, check-in/out",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "HK",
                Name:        "Housekeeping",
                Description: "Kebersihan dan perawatan kamar serta area hotel",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
				ID:          primitive.NewObjectID(),
				Code:        "IT",
				Name:        "Information Technology",
				Description: "Sistem informasi, jaringan, dan support IT",
				Icon:        "💻",
				TotalStaff:  0,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "FB",
                Name:        "Food & Beverage",
                Description: "Layanan makanan dan minuman (restaurant, bar, room service)",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "ENG",
                Name:        "Engineering",
                Description: "Maintenance dan perbaikan fasilitas hotel",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "SEC",
                Name:        "Security",
                Description: "Keamanan hotel dan tamu",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "HR",
                Name:        "Human Resources",
                Description: "Manajemen SDM dan administrasi karyawan",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "FIN",
                Name:        "Finance & Accounting",
                Description: "Keuangan, akuntansi, dan audit",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "SALES",
                Name:        "Sales & Marketing",
                Description: "Penjualan kamar, event, dan promosi hotel",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
            models.Department{
                ID:          primitive.NewObjectID(),
                Code:        "SPA",
                Name:        "Spa & Recreation",
                Description: "Layanan spa, gym, dan fasilitas rekreasi",
                IsActive:    true,
                CreatedAt:   time.Now(),
                UpdatedAt:   time.Now(),
            },
        }

        _, err = collection.InsertMany(ctx, departments)
        return err
    }

    down := func(db *mongo.Database) error {
        ctx := context.Background()
        return db.Collection("departments").Drop(ctx)
    }

    return version, name, description, up, down
}