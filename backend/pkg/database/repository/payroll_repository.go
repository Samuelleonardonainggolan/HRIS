package repository

import (
	"context"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type PayrollRepository interface {
	Create(ctx context.Context, payroll *models.Payroll) error
	Update(ctx context.Context, id primitive.ObjectID, payroll *models.Payroll) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Payroll, error)
	FindAll(ctx context.Context, filter bson.M) ([]models.Payroll, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type payrollRepository struct {
	collection *mongo.Collection
}

func NewPayrollRepository(db *mongo.Database) PayrollRepository {
	return &payrollRepository{
		collection: db.Collection("payrolls"),
	}
}

func (r *payrollRepository) Create(ctx context.Context, payroll *models.Payroll) error {
	payroll.CreatedAt = time.Now()
	payroll.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, payroll)
	return err
}

func (r *payrollRepository) Update(ctx context.Context, id primitive.ObjectID, payroll *models.Payroll) error {
	payroll.UpdatedAt = time.Now()
	update := bson.M{
		"$set": payroll,
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *payrollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Payroll, error) {
	var payroll models.Payroll
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payroll)
	if err != nil {
		return nil, err
	}
	return &payroll, nil
}

func (r *payrollRepository) FindAll(ctx context.Context, filter bson.M) ([]models.Payroll, error) {
	var payrolls []models.Payroll
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &payrolls); err != nil {
		return nil, err
	}

	return payrolls, nil
}

func (r *payrollRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
