// pkg/migration/manager.go
package migration

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Migration struct {
	Version     int
	Name        string
	Description string
	Up          func(db *mongo.Database) error
	Down        func(db *mongo.Database) error
}

type MigrationRecord struct {
	Version   int       `bson:"version"`
	Name      string    `bson:"name"`
	AppliedAt time.Time `bson:"applied_at"`
}

type Manager struct {
	db         *mongo.Database
	migrations []Migration
}

func NewManager(db *mongo.Database) *Manager {
	return &Manager{
		db:         db,
		migrations: []Migration{},
	}
}

func (m *Manager) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// Up - Run all pending migrations
func (m *Manager) Up() error {
	ctx := context.Background()
	collection := m.db.Collection("_migrations")

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	log.Println("🚀 Running migrations...")

	for _, migration := range m.migrations {
		// Check if migration already applied
		var record MigrationRecord
		err := collection.FindOne(ctx, bson.M{"version": migration.Version}).Decode(&record)

		if err == nil {
			log.Printf("⏭️  Skipping migration %d: %s (already applied)", migration.Version, migration.Name)
			continue
		}

		if err != mongo.ErrNoDocuments {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		// Run migration
		log.Printf("⬆️  Running migration %d: %s", migration.Version, migration.Name)

		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %d failed: %v", migration.Version, err)
		}

		// Record migration
		_, err = collection.InsertOne(ctx, MigrationRecord{
			Version:   migration.Version,
			Name:      migration.Name,
			AppliedAt: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to record migration: %v", err)
		}

		log.Printf("✅ Migration %d completed: %s", migration.Version, migration.Name)
	}

	log.Println("🎉 All migrations completed!")
	return nil
}

// Down - Rollback last migration
func (m *Manager) Down() error {
	ctx := context.Background()
	collection := m.db.Collection("_migrations")

	// Get last applied migration
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})
	var record MigrationRecord
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&record)

	if err == mongo.ErrNoDocuments {
		log.Println("ℹ️  No migrations to rollback")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find last migration: %v", err)
	}

	// Find migration
	var targetMigration *Migration
	for _, m := range m.migrations {
		if m.Version == record.Version {
			targetMigration = &m
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found in code", record.Version)
	}

	// Run down
	log.Printf("⬇️  Rolling back migration %d: %s", record.Version, record.Name)

	if err := targetMigration.Down(m.db); err != nil {
		return fmt.Errorf("rollback failed: %v", err)
	}

	// Remove record
	_, err = collection.DeleteOne(ctx, bson.M{"version": record.Version})
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %v", err)
	}

	log.Printf("✅ Rollback completed: %s", record.Name)
	return nil
}

// Status - Show migration status
func (m *Manager) Status() error {
	ctx := context.Background()
	collection := m.db.Collection("_migrations")

	// Get applied migrations
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "version", Value: 1}}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var applied []MigrationRecord
	if err = cursor.All(ctx, &applied); err != nil {
		return err
	}

	appliedMap := make(map[int]MigrationRecord)
	for _, a := range applied {
		appliedMap[a.Version] = a
	}

	// Sort migrations
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	log.Println("\n📊 Migration Status:")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, migration := range m.migrations {
		record, applied := appliedMap[migration.Version]
		if applied {
			log.Printf("✅ [APPLIED] %d: %s (Applied: %s)",
				migration.Version,
				migration.Name,
				record.AppliedAt.Format("2006-01-02 15:04:05"))
		} else {
			log.Printf("⏸️  [PENDING] %d: %s", migration.Version, migration.Name)
		}
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Total migrations: %d | Applied: %d | Pending: %d\n",
		len(m.migrations),
		len(applied),
		len(m.migrations)-len(applied))

	return nil
}

// Reset - Rollback all migrations
func (m *Manager) Reset() error {
	ctx := context.Background()
	collection := m.db.Collection("_migrations")

	log.Println("🔄 Resetting all migrations...")

	// Get all applied migrations
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "version", Value: -1}}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var applied []MigrationRecord
	if err = cursor.All(ctx, &applied); err != nil {
		return err
	}

	// Rollback all
	for _, record := range applied {
		var targetMigration *Migration
		for _, m := range m.migrations {
			if m.Version == record.Version {
				targetMigration = &m
				break
			}
		}

		if targetMigration == nil {
			log.Printf("⚠️  Migration %d not found, skipping rollback", record.Version)
			continue
		}

		log.Printf("⬇️  Rolling back migration %d: %s", record.Version, record.Name)

		if err := targetMigration.Down(m.db); err != nil {
			return fmt.Errorf("rollback failed: %v", err)
		}
	}

	// Clear migration records
	_, err = collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return err
	}

	log.Println("✅ All migrations rolled back!")
	return nil
}

// Fresh - Reset and re-run all migrations
func (m *Manager) Fresh() error {
	if err := m.Reset(); err != nil {
		return err
	}
	return m.Up()
}
