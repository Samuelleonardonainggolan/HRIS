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
	UpdateKepalaDepartemenDecision(ctx context.Context, id string, kepalaID primitive.ObjectID, status string, finalStatus string, rejectionReason string) (*models.OvertimeRequest, error)
	UpdateManagerHRDecision(ctx context.Context, id string, managerHRID primitive.ObjectID, status string, finalStatus string, rejectionReason string) (*models.OvertimeRequest, error)
	Update(ctx context.Context, id string, userID primitive.ObjectID, updates bson.M) (*models.OvertimeRequest, error)
	Delete(ctx context.Context, id string, userID primitive.ObjectID) error
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

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
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
	if req.StatusKepalaDepartemen == "" {
		req.StatusKepalaDepartemen = models.StatusPending
	}
	if req.StatusManagerHR == "" {
		req.StatusManagerHR = models.StatusPending
	}
	if req.FinalStatus == "" {
		req.FinalStatus = models.StatusPending
	}

	result, err := r.collection.InsertOne(ctx, req)
	if err != nil {
		return nil, err
	}

	req.ID = result.InsertedID.(primitive.ObjectID)
	return req, nil
}

func (r *overtimeRequestRepository) UpdateKepalaDepartemenDecision(
	ctx context.Context,
	id string,
	kepalaID primitive.ObjectID,
	status string,
	finalStatus string,
	rejectionReason string,
) (*models.OvertimeRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid overtime request ID")
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "status_kepala_departemen": models.StatusPending}
	setFields := bson.M{
		"kepala_departemen_id":     kepalaID,
		"status_kepala_departemen": status,
		"final_status":             finalStatus,
		"updated_at":               now,
	}
	if status == models.StatusRejected && rejectionReason != "" {
		setFields["rejection_reason_kepala_dept"] = rejectionReason
	}
	update := bson.M{"$set": setFields}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.OvertimeRequest
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan lembur sudah diproses atau bukan wewenang Anda")
		}
		return nil, err
	}

	return &updated, nil
}

func (r *overtimeRequestRepository) UpdateManagerHRDecision(
	ctx context.Context,
	id string,
	managerHRID primitive.ObjectID,
	status string,
	finalStatus string,
	rejectionReason string,
) (*models.OvertimeRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid overtime request ID")
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "status_manager_hr": models.StatusPending}
	setFields := bson.M{
		"manager_hr_id":     managerHRID,
		"status_manager_hr": status,
		"final_status":      finalStatus,
		"updated_at":        now,
	}
	if status == models.StatusRejected && rejectionReason != "" {
		setFields["rejection_reason_manager_hr"] = rejectionReason
	}
	update := bson.M{"$set": setFields}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.OvertimeRequest
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan lembur sudah diproses")
		}
		return nil, err
	}

	return &updated, nil
}

func (r *overtimeRequestRepository) Update(ctx context.Context, id string, userID primitive.ObjectID, updates bson.M) (*models.OvertimeRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid overtime request ID")
	}

	filter := bson.M{
		"_id":     objectID,
		"user_id": userID,
		"final_status": models.StatusPending,
	}

	updates["updated_at"] = time.Now()
	update := bson.M{"$set": updates}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.OvertimeRequest
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan lembur tidak ditemukan atau sudah diproses")
		}
		return nil, err
	}

	return &updated, nil
}

func (r *overtimeRequestRepository) Delete(ctx context.Context, id string, userID primitive.ObjectID) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid overtime request ID")
	}

	filter := bson.M{
		"_id":     objectID,
		"user_id": userID,
		"final_status": models.StatusPending,
	}

	update := bson.M{
		"$set": bson.M{
			"final_status":             "CANCELLED",
			"status_kepala_departemen": "CANCELLED",
			"status_manager_hr":        "CANCELLED",
			"updated_at":               time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("pengajuan lembur tidak ditemukan atau sudah diproses")
	}

	return nil
}
