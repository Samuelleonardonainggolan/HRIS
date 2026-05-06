package repository

import (
	"context"
	"errors"
	"time"

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
	Update(ctx context.Context, id string, req *models.UpdatePositionRequest) error
	FindByName(ctx context.Context, name string) (*models.Position, error)
	FindByNameAndDepartment(ctx context.Context, name string, departmentID string) (*models.Position, error)
	Create(ctx context.Context, position *models.Position) error
	Delete(ctx context.Context, id string) error
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

func (r *positionRepository) Create(ctx context.Context, position *models.Position) error {
	position.ID = primitive.NewObjectID()
	position.CreatedAt = time.Now()
	position.UpdatedAt = time.Now()
	position.IsActive = true

	_, err := r.collection.InsertOne(ctx, position)
	return err
}

func (r *positionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid position ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("position not found")
	}
	return nil
}

func (r *positionRepository) FindByName(ctx context.Context, name string) (*models.Position, error) {
	var position models.Position
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&position)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &position, nil
}

func (r *positionRepository) FindByNameAndDepartment(ctx context.Context, name string, departmentID string) (*models.Position, error) {
	deptOID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	var position models.Position
	err = r.collection.FindOne(ctx, bson.M{"name": name, "department_id": deptOID}).Decode(&position)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &position, nil
}

func (r *positionRepository) Update(ctx context.Context, id string, req *models.UpdatePositionRequest) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid position ID")
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.Code != "" {
		update["$set"].(bson.M)["code"] = req.Code
	}
	if req.Name != "" {
		update["$set"].(bson.M)["name"] = req.Name
	}
	if req.DepartmentID != "" {
		deptOID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err != nil {
			return errors.New("invalid department ID")
		}
		update["$set"].(bson.M)["department_id"] = deptOID
	}
	if req.Level != 0 {
		update["$set"].(bson.M)["level"] = req.Level
	}
	if req.Description != "" {
		update["$set"].(bson.M)["description"] = req.Description
	}
	if req.Requirements != "" {
		update["$set"].(bson.M)["requirements"] = req.Requirements
	}
	if req.SalaryRange != nil {
		update["$set"].(bson.M)["salary_range"] = *req.SalaryRange
	}
	if req.IsActive != nil {
		update["$set"].(bson.M)["is_active"] = *req.IsActive
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("position not found")
	}
	return nil
}
