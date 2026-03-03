// pkg/database/mongodb.go
package database

import (
    "context"
    "fmt"
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
    // Buat context dengan timeout lebih lama
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    // Set client options dengan timeout lebih besar
    clientOptions := options.Client().
        ApplyURI(uri).
        SetServerSelectionTimeout(60 * time.Second).
        SetConnectTimeout(60 * time.Second).
        SetSocketTimeout(60 * time.Second).
        SetMaxConnIdleTime(60 * time.Second).
        SetMaxPoolSize(50).
        SetMinPoolSize(10)

    log.Println("⏳ Connecting to MongoDB Atlas (this may take up to 60 seconds)...")

    // Connect to MongoDB
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
    }

    // Ping the database dengan context yang lebih lama
    pingCtx, pingCancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer pingCancel()

    log.Println("🏓 Pinging MongoDB...")
    if err := client.Ping(pingCtx, nil); err != nil {
        return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
    }

    log.Println("✅ Connected to MongoDB successfully")

    db := client.Database(dbName)

    // Create indexes
    if err := createIndexes(ctx, db); err != nil {
        return nil, fmt.Errorf("failed to create indexes: %v", err)
    }

    return &MongoDB{
        Client:   client,
        Database: db,
    }, nil
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
    userCollection := db.Collection("users")

    // Unique index on email
    emailIndexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    }

    // Unique index on nik
    nikIndexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "nik", Value: 1}},
        Options: options.Index().SetUnique(true),
    }

    _, err := userCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{emailIndexModel, nikIndexModel})
    if err != nil {
        return err
    }

    log.Println("✅ Indexes created successfully")
    return nil
}

func (m *MongoDB) Disconnect() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return m.Client.Disconnect(ctx)
}