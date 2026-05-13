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

type OvertimeRequestRepository interface {
	FindByID(ctx context.Context, id string) (*models.OvertimeRequest, error)
	Find(ctx context.Context, filter bson.M) ([]models.OvertimeRequest, error)
	Create(ctx context.Context, req *models.OvertimeRequest) (*models.OvertimeRequest, error)
	Update(ctx context.Context, id string, updates bson.M) (*models.OvertimeRequest, error)
	Delete(ctx context.Context, id string) error
	UpdateEmployeeStatus(ctx context.Context, overtimeID string, userID string, status string, rejectionNote string) error
	UpdateEmployeeLetterURL(ctx context.Context, overtimeID string, userID string, letterURL string) error
	UpdateEmployeeReward(ctx context.Context, overtimeID string, userID string, reward models.OvertimeReward) error
}

type overtimeRequestRepository struct {
	collection *mongo.Collection
}

func NewOvertimeRequestRepository(db *mongo.Database) OvertimeRequestRepository {
	return &overtimeRequestRepository{collection: db.Collection("overtime_requests")}
}

func (r *overtimeRequestRepository) FindByID(ctx context.Context, id string) (*models.OvertimeRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid overtime request ID")
	}

	var req models.OvertimeRequest
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&req)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("overtime request not found")
		}
		return nil, err
	}

	return &req, nil
}

func (r *overtimeRequestRepository) Find(ctx context.Context, filter bson.M) ([]models.OvertimeRequest, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}, {Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.OvertimeRequest
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *overtimeRequestRepository) Create(ctx context.Context, req *models.OvertimeRequest) (*models.OvertimeRequest, error) {
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now
	if req.Status == "" {
		req.Status = models.StatusDraft
	}

	result, err := r.collection.InsertOne(ctx, req)
	if err != nil {
		return nil, err
	}

	req.ID = result.InsertedID.(primitive.ObjectID)
	return req, nil
}

func (r *overtimeRequestRepository) Update(ctx context.Context, id string, updates bson.M) (*models.OvertimeRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid overtime request ID")
	}

	updates["updated_at"] = time.Now()
	update := bson.M{"$set": updates}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.OvertimeRequest
	err = r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan lembur tidak ditemukan")
		}
		return nil, err
	}

	return &updated, nil
}

func (r *overtimeRequestRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid overtime request ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID, "status": models.StatusDraft})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("pengajuan lembur tidak ditemukan atau tidak dapat dihapus (sudah submitted/published)")
	}

	return nil
}

func (r *overtimeRequestRepository) UpdateEmployeeStatus(ctx context.Context, overtimeID string, userID string, status string, rejectionNote string) error {
	oid, err := primitive.ObjectIDFromHex(overtimeID)
	if err != nil {
		return errors.New("invalid overtime ID")
	}
	uoid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	filter := bson.M{
		"_id":               oid,
		"employees.user_id": uoid,
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"employees.$.employee_status": status,
			"employees.$.rejection_note":  rejectionNote,
			"employees.$.confirmed_at":    &now,
			"updated_at":                 now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("karyawan tidak ditemukan dalam pengajuan lembur ini")
	}

	return nil
}
func (r *overtimeRequestRepository) UpdateEmployeeLetterURL(ctx context.Context, overtimeID string, userID string, letterURL string) error {
	oid, err := primitive.ObjectIDFromHex(overtimeID)
	if err != nil {
		return errors.New("invalid overtime ID")
	}
	uoid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	filter := bson.M{
		"_id":               oid,
		"employees.user_id": uoid,
	}

	update := bson.M{
		"$set": bson.M{
			"employees.$.letter_url": letterURL,
			"updated_at":           time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("karyawan tidak ditemukan dalam pengajuan lembur ini")
	}

	return nil
}

func (r *overtimeRequestRepository) UpdateEmployeeReward(ctx context.Context, overtimeID string, userID string, reward models.OvertimeReward) error {
	oid, err := primitive.ObjectIDFromHex(overtimeID)
	if err != nil {
		return errors.New("invalid overtime ID")
	}
	uoid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	filter := bson.M{
		"_id":               oid,
		"employees.user_id": uoid,
	}

	update := bson.M{
		"$set": bson.M{
			"employees.$.reward": reward,
			"updated_at":        time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("karyawan tidak ditemukan dalam pengajuan lembur ini")
	}

	return nil
}
