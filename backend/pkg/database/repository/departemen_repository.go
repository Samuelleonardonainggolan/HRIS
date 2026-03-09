// pkg/database/repository/department_repository.go
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

type DepartmentRepository interface {
	Create(ctx context.Context, department *models.Department) error
	FindByID(ctx context.Context, id string) (*models.Department, error)
	FindByCode(ctx context.Context, code string) (*models.Department, error)
	FindByName(ctx context.Context, name string) (*models.Department, error)
	FindAll(ctx context.Context) ([]models.Department, error)
	Update(ctx context.Context, id string, department *models.UpdateDepartmentRequest) error
	Delete(ctx context.Context, id string) error
}

type departmentRepository struct {
	collection *mongo.Collection
}

func NewDepartmentRepository(db *mongo.Database) DepartmentRepository {
	return &departmentRepository{
		collection: db.Collection("departments"),
	}
}

func (r *departmentRepository) Create(ctx context.Context, department *models.Department) error {
	department.ID = primitive.NewObjectID()
	department.CreatedAt = time.Now()
	department.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, department)
	return err
}

func (r *departmentRepository) FindByID(ctx context.Context, id string) (*models.Department, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	var department models.Department
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&department)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("department not found")
		}
		return nil, err
	}
	return &department, nil
}

func (r *departmentRepository) FindByCode(ctx context.Context, code string) (*models.Department, error) {
	var department models.Department
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&department)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &department, nil
}

func (r *departmentRepository) FindByName(ctx context.Context, name string) (*models.Department, error) {
	var department models.Department
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&department)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &department, nil
}

func (r *departmentRepository) FindAll(ctx context.Context) ([]models.Department, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var departments []models.Department
	if err = cursor.All(ctx, &departments); err != nil {
		return nil, err
	}
	return departments, nil
}

func (r *departmentRepository) Update(ctx context.Context, id string, req *models.UpdateDepartmentRequest) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid department ID")
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
	if req.Description != "" {
		update["$set"].(bson.M)["description"] = req.Description
	}
	if req.Icon != "" {
		update["$set"].(bson.M)["icon"] = req.Icon
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("department not found")
	}
	return nil
}

func (r *departmentRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid department ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("department not found")
	}
	return nil
}