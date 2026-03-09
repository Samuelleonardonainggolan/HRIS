// pkg/database/repository/user_repository.go
package repository

import (
	"context"
	"errors"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByPayrollNumber(ctx context.Context, payrollNumber string) (*models.User, error) // Changed from FindByNIK
	FindAll(ctx context.Context) ([]models.User, error)
	FindByDepartment(ctx context.Context, departmentID string) ([]models.User, error)
	Update(ctx context.Context, id string, user *models.UpdateUserRequest) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Changed from FindByNIK to FindByPayrollNumber
func (r *userRepository) FindByPayrollNumber(ctx context.Context, payrollNumber string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"payroll_number": payrollNumber}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) FindByDepartment(ctx context.Context, departmentID string) ([]models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"department_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id string, req *models.UpdateUserRequest) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	update := bson.M{"$set": bson.M{}}

	// Only update provided fields
	if req.PayrollNumber != "" { // Changed from NIK
		update["$set"].(bson.M)["payroll_number"] = req.PayrollNumber
	}
	if req.FullName != "" {
		update["$set"].(bson.M)["full_name"] = req.FullName
	}
	if req.BirthDate != "" {
		update["$set"].(bson.M)["birth_date"] = req.BirthDate
	}
	if req.Religion != "" {
		update["$set"].(bson.M)["religion"] = req.Religion
	}
	if req.LastEducation != "" {
		update["$set"].(bson.M)["last_education"] = req.LastEducation
	}
	if req.YearEnrolled != "" {
		update["$set"].(bson.M)["year_enrolled"] = req.YearEnrolled
	}
	if req.EmploymentStatus != "" {
		update["$set"].(bson.M)["employment_status"] = req.EmploymentStatus
	}
	if req.DepartmentID != "" {
		deptOID, _ := primitive.ObjectIDFromHex(req.DepartmentID)
		update["$set"].(bson.M)["department_id"] = deptOID
	}
	if req.PositionID != "" {
		posOID, _ := primitive.ObjectIDFromHex(req.PositionID)
		update["$set"].(bson.M)["position_id"] = posOID
	}
	if req.Phone != "" {
		update["$set"].(bson.M)["phone"] = req.Phone
	}
	if req.Address != "" {
		update["$set"].(bson.M)["address"] = req.Address
	}
	if req.IsActive != nil {
		update["$set"].(bson.M)["is_active"] = *req.IsActive
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}