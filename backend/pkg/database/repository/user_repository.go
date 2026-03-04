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
    GetAll(ctx context.Context) ([]models.User, error)
    Update(ctx context.Context, id string, user *models.User) error
    Delete(ctx context.Context, id string) error
    UpdateProfile(ctx context.Context, id string, req models.UpdateProfileRequest) error
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

func (r *userRepository) GetAll(ctx context.Context) ([]models.User, error) {
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

func (r *userRepository) Update(ctx context.Context, id string, user *models.User) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return errors.New("invalid user ID")
    }

    user.UpdatedAt = time.Now()

    update := bson.M{
        "$set": user,
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

func (r *userRepository) UpdateProfile(ctx context.Context, id string, req models.UpdateProfileRequest) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return errors.New("invalid user ID")
    }

    update := bson.M{
        "$set": bson.M{
            "updated_at": time.Now(),
        },
    }

    // Only update fields that are provided
    setFields := update["$set"].(bson.M)
    
    if req.FullName != "" {
        setFields["full_name"] = req.FullName
    }
    if req.Phone != "" {
        setFields["phone"] = req.Phone
    }
    if req.Address != "" {
        setFields["address"] = req.Address
    }
    if req.Avatar != "" {
        setFields["avatar"] = req.Avatar
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