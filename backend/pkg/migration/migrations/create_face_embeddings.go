// pkg/migration/migrations/004_create_face_embeddings.go
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

func CreateFaceEmbeddings() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
    version := 4
    name := "create_face_embeddings"
    description := "Create face_embeddings collection with indexes"

    up := func(db *mongo.Database) error {
        ctx := context.Background()
        collection := db.Collection("face_embeddings")

        // Create indexes
        indexModels := []mongo.IndexModel{
            {
                // Unique index on user_id (one face per user)
                Keys:    bson.D{{Key: "user_id", Value: 1}},
                Options: options.Index().SetUnique(true),
            },
            {
                // Index on is_active for quick filtering
                Keys: bson.D{{Key: "is_active", Value: 1}},
            },
            {
                // Compound index for active face lookups
                Keys: bson.D{
                    {Key: "user_id", Value: 1},
                    {Key: "is_active", Value: 1},
                },
            },
        }

        _, err := collection.Indexes().CreateMany(ctx, indexModels)
        if err != nil {
            return err
        }

        // Optional: Create sample face embedding for test user
        // Get Manager HR user
        userCollection := db.Collection("users")
        var user models.User
        err = userCollection.FindOne(ctx, bson.M{"email": "manager.hr@company.com"}).Decode(&user)
        if err != nil {
            // If user not found, skip creating sample embedding
            return nil
        }

        // Create sample face embedding (random vector - in production this comes from face recognition model)
        sampleEmbedding := make([]float32, 128) // 128-dimensional vector (common for face recognition)
        for i := range sampleEmbedding {
            sampleEmbedding[i] = 0.0 // Initialize with zeros (in production, use actual embedding)
        }

        faceEmbedding := models.FaceEmbedding{
            ID:              primitive.NewObjectID(),
            UserID:          user.ID,
            FaceEmbedding:   sampleEmbedding,
            FaceImageURL:    "", // Will be set when user uploads photo
            IsActive:        true,
            RegisteredAt:    time.Now(),
            CreatedAt:       time.Now(),
            UpdatedAt:       time.Now(),
        }

        _, err = collection.InsertOne(ctx, faceEmbedding)
        return err
    }

    down := func(db *mongo.Database) error {
        ctx := context.Background()
        return db.Collection("face_embeddings").Drop(ctx)
    }

    return version, name, description, up, down
}
