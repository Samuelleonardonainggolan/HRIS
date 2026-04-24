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

type PengajuanIzinCutiRepository interface {
	FindByID(ctx context.Context, id string) (*models.LeaveRequest, error)
	Find(ctx context.Context, filter bson.M) ([]models.LeaveRequest, error)
	UpdateManagerHRDecision(ctx context.Context, id string, managerHRID primitive.ObjectID, statusManagerHR string, finalStatus string, rejectionReason string) (*models.LeaveRequest, error)
	UpdateKepalaDepartemenDecision(ctx context.Context, id string, kepalaDepartemenID primitive.ObjectID, statusKepalaDepartemen string, finalStatus string, rejectionReason string) (*models.LeaveRequest, error)
}

type pengajuanIzinCutiRepository struct {
	collection *mongo.Collection
}

func NewPengajuanIzinCutiRepository(db *mongo.Database) PengajuanIzinCutiRepository {
	return &pengajuanIzinCutiRepository{collection: db.Collection("leave_request")} // ✅ renamed
}

func (r *pengajuanIzinCutiRepository) FindByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid pengajuan ID")
	}

	var pengajuan models.LeaveRequest
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&pengajuan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan not found")
		}
		return nil, err
	}

	return &pengajuan, nil
}

func (r *pengajuanIzinCutiRepository) Find(ctx context.Context, filter bson.M) ([]models.LeaveRequest, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pengajuans []models.LeaveRequest
	if err := cursor.All(ctx, &pengajuans); err != nil {
		return nil, err
	}

	return pengajuans, nil
}

func (r *pengajuanIzinCutiRepository) UpdateManagerHRDecision(
	ctx context.Context,
	id string,
	managerHRID primitive.ObjectID,
	statusManagerHR string,
	finalStatus string,
	rejectionReason string,
) (*models.LeaveRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid pengajuan ID")
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "status_manager_hr": models.StatusPending}
	setFields := bson.M{
		"manager_hr_id":     managerHRID,
		"status_manager_hr": statusManagerHR,
		"final_status":      finalStatus,
		"updated_at":        now,
	}
	if statusManagerHR == models.StatusRejected && rejectionReason != "" {
		setFields["rejection_reason_manager_hr"] = rejectionReason
	}
	update := bson.M{"$set": setFields}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.LeaveRequest
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan sudah diproses")
		}
		return nil, err
	}

	return &updated, nil
}

func (r *pengajuanIzinCutiRepository) UpdateKepalaDepartemenDecision(
	ctx context.Context,
	id string,
	kepalaDepartemenID primitive.ObjectID,
	statusKepalaDepartemen string,
	finalStatus string,
	rejectionReason string,
) (*models.LeaveRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid pengajuan ID")
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "status_kepala_departemen": models.StatusPending}
	setFields := bson.M{
		"kepala_departemen_id":     kepalaDepartemenID,
		"status_kepala_departemen": statusKepalaDepartemen,
		"final_status":             finalStatus,
		"updated_at":               now,
	}
	if statusKepalaDepartemen == models.StatusRejected && rejectionReason != "" {
		setFields["rejection_reason_kepala_dept"] = rejectionReason
	}
	update := bson.M{"$set": setFields}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.LeaveRequest
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan sudah diproses atau bukan wewenang Anda")
		}
		return nil, err
	}

	return &updated, nil
}
