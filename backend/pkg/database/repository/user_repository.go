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
    "go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    FindByNIK(ctx context.Context, nik string) (*models.User, error)
    FindByID(ctx context.Context, id string) (*models.User, error)
    Update(ctx context.Context, id string, update *models.UpdateProfileRequest) error
    Delete(ctx context.Context, id string) error
    GetAll(ctx context.Context) ([]*models.User, error)
    GetByDepartment(ctx context.Context, department string) ([]*models.User, error)
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
    user.ID = primitive.NewObjectID()
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    user.IsActive = true

    _, err := r.collection.InsertOne(ctx, user)
    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            return errors.New("email or NIK already exists")
        }
        return err
    }

    return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User

    filter := bson.M{
        "email":     email,
        "is_active": true,
    }

    err := r.collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, err
    }

    return &user, nil
}

func (r *userRepository) FindByNIK(ctx context.Context, nik string) (*models.User, error) {
    var user models.User

    filter := bson.M{
        "nik":       nik,
        "is_active": true,
    }

    err := r.collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
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

    filter := bson.M{
        "_id":       objectID,
        "is_active": true,
    }

    err = r.collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, err
    }

    return &user, nil
}

func (r *userRepository) Update(ctx context.Context, id string, updateReq *models.UpdateProfileRequest) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return errors.New("invalid user ID")
    }

    update := bson.M{
        "$set": bson.M{
            "full_name":  updateReq.FullName,
            "phone":      updateReq.Phone,
            "address":    updateReq.Address,
            "avatar":     updateReq.Avatar,
            "department": updateReq.Department,
            "position":   updateReq.Position,
            "updated_at": time.Now(),
        },
    }

    filter := bson.M{"_id": objectID}

    result, err := r.collection.UpdateOne(ctx, filter, update)
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

    update := bson.M{
        "$set": bson.M{
            "is_active":  false,
            "updated_at": time.Now(),
        },
    }

    filter := bson.M{"_id": objectID}

    result, err := r.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }

    if result.MatchedCount == 0 {
        return errors.New("user not found")
    }

    return nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*models.User, error) {
    filter := bson.M{"is_active": true}

    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var users []*models.User
    if err = cursor.All(ctx, &users); err != nil {
        return nil, err
    }

    return users, nil
}

func (r *userRepository) GetByDepartment(ctx context.Context, department string) ([]*models.User, error) {
    filter := bson.M{
        "department": department,
        "is_active":  true,
    }

    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var users []*models.User
    if err = cursor.All(ctx, &users); err != nil {
        return nil, err
    }

    return users, nil
}