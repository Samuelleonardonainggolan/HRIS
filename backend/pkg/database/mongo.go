// pkg/database/mongodb.go
package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	log.Println("⏳ Connecting to MongoDB Atlas (this may take up to 60 seconds)...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri).
		SetServerSelectionTimeout(30 * time.Second).
		SetConnectTimeout(30 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	log.Println("🏓 Pinging MongoDB...")
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to MongoDB successfully")

	database := client.Database(dbName)

	// Clean up old indexes
	if err := cleanupOldIndexes(database); err != nil {
		log.Printf("⚠️  Warning: Could not clean up old indexes: %v", err)
	}

	// Create required indexes
	if err := createIndexes(database); err != nil {
		return nil, err
	}

	log.Println("✅ Indexes created successfully")

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

// cleanupOldIndexes removes deprecated indexes
func cleanupOldIndexes(db *mongo.Database) error {
	ctx := context.Background()

	// List of old indexes to drop
	oldIndexes := map[string][]string{
		"users": {"nik_1"}, // Old NIK index
	}

	for collectionName, indexes := range oldIndexes {
		collection := db.Collection(collectionName)

		for _, indexName := range indexes {
			// ✅ Fix: Handle both return values
			_, err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				// Ignore "index not found" errors
				if mongo.IsNetworkError(err) || mongo.IsTimeout(err) {
					return err
				}
				log.Printf("Note: Could not drop index %s from %s (might not exist): %v", indexName, collectionName, err)
			} else {
				log.Printf("✅ Dropped old index: %s from %s", indexName, collectionName)
			}
		}
	}

	return nil
}

func createIndexes(db *mongo.Database) error {
	ctx := context.Background()

	// Users collection indexes
	userCollection := db.Collection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "payroll_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
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

	_, err := userCollection.Indexes().CreateMany(ctx, userIndexes)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}
