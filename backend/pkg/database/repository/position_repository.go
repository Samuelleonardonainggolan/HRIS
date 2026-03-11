package repository

import (
	"context"
	"errors"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PositionRepository interface {
	FindByID(ctx context.Context, id string) (*models.Position, error)
	FindAll(ctx context.Context) ([]models.Position, error)
	FindByDepartment(ctx context.Context, departmentID string) ([]models.Position, error)
}

type positionRepository struct {
	collection *mongo.Collection
}

func NewPositionRepository(db *mongo.Database) PositionRepository {
	return &positionRepository{
		collection: db.Collection("positions"),
	}
}

func (r *positionRepository) FindByID(ctx context.Context, id string) (*models.Position, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid position ID")
	}

	var position models.Position
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&position)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("position not found")
		}
		return nil, err
	}
	return &position, nil
}

func (r *positionRepository) FindAll(ctx context.Context) ([]models.Position, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var positions []models.Position
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *positionRepository) FindByDepartment(ctx context.Context, departmentID string) ([]models.Position, error) {
	deptOID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	opts := options.Find().SetSort(bson.D{{Key: "level", Value: 1}, {Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"department_id": deptOID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var positions []models.Position
	if err := cursor.All(ctx, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}
