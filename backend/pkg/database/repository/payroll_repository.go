// pkg/database/repository/payroll_repository.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PayrollRepository interface {
	FindByUserAndMonthYear(ctx context.Context, userID string, month, year int) (*models.Payroll, error)
	Create(ctx context.Context, payroll *models.Payroll) error
}

type payrollRepository struct {
	collection *mongo.Collection
}

func NewPayrollRepository(db *mongo.Database) PayrollRepository {
	return &payrollRepository{
		collection: db.Collection("payrolls"),
	}
}

func (r *payrollRepository) FindByUserAndMonthYear(ctx context.Context, userID string, month, year int) (*models.Payroll, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var payroll models.Payroll
	err = r.collection.FindOne(ctx, bson.M{
		"user_id": objectID,
		"month":   month,
		"year":    year,
	}).Decode(&payroll)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &payroll, nil
}

func (r *payrollRepository) Create(ctx context.Context, payroll *models.Payroll) error {
	if payroll.ID.IsZero() {
		payroll.ID = primitive.NewObjectID()
	}
	now := time.Now()
	if payroll.CreatedAt.IsZero() {
		payroll.CreatedAt = now
	}
	payroll.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, payroll)
	return err
}
