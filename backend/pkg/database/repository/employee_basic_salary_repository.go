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

type EmployeeBasicSalaryRepository interface {
	Create(ctx context.Context, salary *models.EmployeeBasicSalary) error
	FindByID(ctx context.Context, id string) (*models.EmployeeBasicSalary, error)

	FindActiveByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalary, error)
	FindAll(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.EmployeeBasicSalary, error)

	UpdateByID(ctx context.Context, id string, req *models.UpdateEmployeeBasicSalaryRequest) error
	UpdateEffectiveFromByID(ctx context.Context, id string, t time.Time) error
	DeactivateByID(ctx context.Context, id string) error
	FindActiveByUserIDs(ctx context.Context, userIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error)
	FindLatestByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalary, error)
}

type employeeBasicSalaryRepository struct {
	collection *mongo.Collection
}

func (r *employeeBasicSalaryRepository) FindLatestByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalary, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	opts := options.FindOne().SetSort(bson.D{
		{Key: "effective_from", Value: -1},
		{Key: "created_at", Value: -1},
	})

	var out models.EmployeeBasicSalary
	err = r.collection.FindOne(ctx, bson.M{"user_id": uid}, opts).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func NewEmployeeBasicSalaryRepository(db *mongo.Database) EmployeeBasicSalaryRepository {
	return &employeeBasicSalaryRepository{
		collection: db.Collection("employee_basic_salaries"),
	}
}

// pkg/database/repository/employee_basic_salary_repository.go
func (r *employeeBasicSalaryRepository) FindActiveByUserIDs(
  ctx context.Context,
  userIDs []primitive.ObjectID,
) (map[primitive.ObjectID]bool, error) {
  out := map[primitive.ObjectID]bool{}
  if len(userIDs) == 0 {
    return out, nil
  }

  cur, err := r.collection.Find(ctx, bson.M{
    "user_id": bson.M{"$in": userIDs},
    "is_active": true,
  }, options.Find().SetProjection(bson.M{"user_id": 1}))
  if err != nil {
    return nil, err
  }
  defer cur.Close(ctx)

  type doc struct {
    UserID primitive.ObjectID `bson:"user_id"`
  }
  for cur.Next(ctx) {
    var d doc
    if err := cur.Decode(&d); err == nil {
      out[d.UserID] = true
    }
  }
  return out, nil
}

func (r *employeeBasicSalaryRepository) Create(ctx context.Context, salary *models.EmployeeBasicSalary) error {
	salary.ID = primitive.NewObjectID()
	salary.CreatedAt = time.Now()
	salary.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, salary)
	return err
}

func (r *employeeBasicSalaryRepository) FindByID(ctx context.Context, id string) (*models.EmployeeBasicSalary, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid salary ID")
	}

	var out models.EmployeeBasicSalary
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("salary not found")
		}
		return nil, err
	}
	return &out, nil
}

func (r *employeeBasicSalaryRepository) FindActiveByUserID(ctx context.Context, userID string) (*models.EmployeeBasicSalary, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var out models.EmployeeBasicSalary
	err = r.collection.FindOne(ctx, bson.M{"user_id": uid, "is_active": true}).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *employeeBasicSalaryRepository) FindAll(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]models.EmployeeBasicSalary, error) {
	cur, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.EmployeeBasicSalary
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *employeeBasicSalaryRepository) UpdateByID(ctx context.Context, id string, req *models.UpdateEmployeeBasicSalaryRequest) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid salary ID")
	}

	set := bson.M{"updated_at": time.Now()}
	if req.BasicSalary != nil {
		set["basic_salary"] = *req.BasicSalary
	}
	if req.IsActive != nil {
		set["is_active"] = *req.IsActive
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("salary not found")
	}
	return nil
}

func (r *employeeBasicSalaryRepository) UpdateEffectiveFromByID(ctx context.Context, id string, t time.Time) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid salary ID")
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$set": bson.M{
			"effective_from": t,
			"updated_at":     time.Now(),
		},
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("salary not found")
	}
	return nil
}

func (r *employeeBasicSalaryRepository) DeactivateByID(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid salary ID")
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("salary not found")
	}
	return nil
}