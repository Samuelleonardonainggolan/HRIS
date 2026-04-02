// pkg/database/repository/user_repository.go
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

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByPayrollNumber(ctx context.Context, payrollNumber string) (*models.User, error) // Changed from FindByNIK
	FindAll(ctx context.Context) ([]models.User, error)
	FindByIDs(ctx context.Context, ids []string) ([]models.User, error)
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
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now
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

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) ([]models.User, error) {
	if len(ids) == 0 {
		return []models.User{}, nil
	}

	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, errors.New("invalid user ID")
		}
		objectIDs = append(objectIDs, oid)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
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

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	// Only update provided fields
	if req.PayrollNumber != "" { // Changed from NIK
		update["$set"].(bson.M)["payroll_number"] = req.PayrollNumber
	}
	if req.FullName != "" {
		update["$set"].(bson.M)["full_name"] = req.FullName
	}
	if req.BirthDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.BirthDate)
		if parseErr != nil {
			parsed, parseErr = time.Parse(time.RFC3339, req.BirthDate)
		}
		if parseErr != nil {
			return errors.New("invalid birth_date format")
		}
		update["$set"].(bson.M)["birth_date"] = parsed
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
		deptOID, deptErr := primitive.ObjectIDFromHex(req.DepartmentID)
		if deptErr != nil {
			return errors.New("invalid department_id")
		}
		update["$set"].(bson.M)["department_id"] = deptOID
		if req.DepartmentName != "" {
			update["$set"].(bson.M)["department_name"] = req.DepartmentName
		}
	}
	if req.PositionID != "" {
		posOID, posErr := primitive.ObjectIDFromHex(req.PositionID)
		if posErr != nil {
			return errors.New("invalid position_id")
		}
		update["$set"].(bson.M)["position_id"] = posOID
		if req.PositionName != "" {
			update["$set"].(bson.M)["position_name"] = req.PositionName
		}
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
