// cmd/migrate/main.go
package main

import (
	"flag"
	"log"
	"os"

	"github.com/andikatampubolon10/hris-backend/internal/config"
	"github.com/andikatampubolon10/hris-backend/pkg/database"
	"github.com/andikatampubolon10/hris-backend/pkg/migration"
	"github.com/andikatampubolon10/hris-backend/pkg/migration/migrations"
)

func main() {
	// Parse command
	upCmd := flag.Bool("up", false, "Run all pending migrations")
	downCmd := flag.Bool("down", false, "Rollback last migration")
	statusCmd := flag.Bool("status", false, "Show migration status")
	resetCmd := flag.Bool("reset", false, "Rollback all migrations")
	freshCmd := flag.Bool("fresh", false, "Reset and re-run all migrations")
	flag.Parse()

	// Load config
	cfg := config.LoadConfig()

	// Connect to MongoDB
	mongodb, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongodb.Disconnect()

	// Create migration manager
	manager := migration.NewManager(mongodb.Database)

	// Register migrations
	registerMigrations(manager)

	// Execute command
	switch {
	case *upCmd:
		if err := manager.Up(); err != nil {
			log.Fatal("Migration failed:", err)
		}

	case *downCmd:
		if err := manager.Down(); err != nil {
			log.Fatal("Rollback failed:", err)
		}

	case *statusCmd:
		if err := manager.Status(); err != nil {
			log.Fatal("Status check failed:", err)
		}

	case *resetCmd:
		if err := manager.Reset(); err != nil {
			log.Fatal("Reset failed:", err)
		}

	case *freshCmd:
		if err := manager.Fresh(); err != nil {
			log.Fatal("Fresh migration failed:", err)
		}

	default:
		log.Println("Usage:")
		log.Println("  go run cmd/migrate/main.go -up        # Run pending migrations")
		log.Println("  go run cmd/migrate/main.go -down      # Rollback last migration")
		log.Println("  go run cmd/migrate/main.go -status    # Show migration status")
		log.Println("  go run cmd/migrate/main.go -reset     # Rollback all migrations")
		log.Println("  go run cmd/migrate/main.go -fresh     # Reset and re-run all")
		os.Exit(1)
	}
}

func registerMigrations(manager *migration.Manager) {
	// Migration 1: Departments
	version, name, desc, up, down := migrations.CreateDepartments()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 2: Positions
	version, name, desc, up, down = migrations.CreatePositions()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 3: Test Users
	version, name, desc, up, down = migrations.CreateTestUsers()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 4: Face Embeddings
	version, name, desc, up, down = migrations.CreateFaceEmbeddings()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 5: Geofences
	version, name, desc, up, down = migrations.CreateGeofences()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 6: Attendances
	version, name, desc, up, down = migrations.CreateAttendances()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 7: Jam Kerja
	version, name, desc, up, down = migrations.CreateJamKerja()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 8: Waktu Istirahat
	version, name, desc, up, down = migrations.CreateBreakTime()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 9: Kategori Pengajuan
	version, name, desc, up, down = migrations.CreateKategoriPengajuan()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 10: Tipe Pengajuan
	version, name, desc, up, down = migrations.CreateTipePengajuan()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 11: Pengajuan Izin/Cuti
	version, name, desc, up, down = migrations.CreatePengajuanIzinCuti()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 12: Leave Balance
	version, name, desc, up, down = migrations.CreateLeaveBalance()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	version, name, desc, up, down = migrations.CreateEmployeeBasicSalaries()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})

	// Migration 12: Migrate legacy pengajuan_izin_cuti
	version, name, desc, up, down = migrations.MigratePengajuanIzinCutiLegacy()
	manager.Register(migration.Migration{
		Version:     version,
		Name:        name,
		Description: desc,
		Up:          up,
		Down:        down,
	})
}
