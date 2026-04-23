package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateFaceRegistrationRequests() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 17
	name := "create_face_registration_requests"
	description := "Create face_registration_requests collection and indexes"

	up := func(db *mongo.Database) error {
		ctx := context.Background()
		collection := db.Collection("face_registration_requests")

		indexes := []mongo.IndexModel{

			{
				Keys: bson.D{{Key: "user_id", Value: 1}},
			},

			{
				Keys: bson.D{{Key: "status", Value: 1}, {Key: "submitted_at", Value: -1}},
			},

			{
				Keys: bson.D{{Key: "face_embedding_id", Value: 1}},
				Options: options.Index().
					SetUnique(true).
					SetPartialFilterExpression(bson.M{"face_embedding_id": bson.M{"$exists": true}}),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexes)
		return err
	}

	down := func(db *mongo.Database) error {
		ctx := context.Background()
		return db.Collection("face_registration_requests").Drop(ctx)
	}

	return version, name, description, up, down
}