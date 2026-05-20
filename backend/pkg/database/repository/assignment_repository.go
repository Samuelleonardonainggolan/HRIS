package repository

import (
	"context"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *models.Assignment) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Assignment, error)
	ListByDepartment(ctx context.Context, departmentID primitive.ObjectID) ([]models.Assignment, error)
	Update(ctx context.Context, assignment *models.Assignment) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	ListByEmployee(ctx context.Context, userID primitive.ObjectID) ([]models.Assignment, error)
	Find(ctx context.Context, filter bson.M) ([]models.Assignment, error)
}

type assignmentRepository struct {
	collection *mongo.Collection
}

func NewAssignmentRepository(db *mongo.Database) AssignmentRepository {
	return &assignmentRepository{
		collection: db.Collection("assignments"),
	}
}

func (r *assignmentRepository) Create(ctx context.Context, assignment *models.Assignment) error {
	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, assignment)
	return err
}

func (r *assignmentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&assignment)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *assignmentRepository) ListByDepartment(ctx context.Context, departmentID primitive.ObjectID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"department_id": departmentID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *assignmentRepository) Update(ctx context.Context, assignment *models.Assignment) error {
	assignment.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": assignment.ID},
		bson.M{"$set": assignment},
	)
	return err
}

func (r *assignmentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *assignmentRepository) ListByEmployee(ctx context.Context, userID primitive.ObjectID) ([]models.Assignment, error) {
	var assignments []models.Assignment
	// Mencari penugasan di mana userID ada di dalam array employees
	filter := bson.M{"employees.user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *assignmentRepository) Find(ctx context.Context, filter bson.M) ([]models.Assignment, error) {
	var assignments []models.Assignment
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}
