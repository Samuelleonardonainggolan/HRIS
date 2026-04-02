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
	FindByID(ctx context.Context, id string) (*models.PengajuanIzinCuti, error)
	Find(ctx context.Context, filter bson.M) ([]models.PengajuanIzinCuti, error)
	UpdateManagerHRDecision(ctx context.Context, id string, managerHRID primitive.ObjectID, statusManagerHR string, statusFinal string) (*models.PengajuanIzinCuti, error)
}

type pengajuanIzinCutiRepository struct {
	collection *mongo.Collection
}

func NewPengajuanIzinCutiRepository(db *mongo.Database) PengajuanIzinCutiRepository {
	return &pengajuanIzinCutiRepository{collection: db.Collection("pengajuan_izin_cuti")}
}

func (r *pengajuanIzinCutiRepository) FindByID(ctx context.Context, id string) (*models.PengajuanIzinCuti, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid pengajuan ID")
	}

	var pengajuan models.PengajuanIzinCuti
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&pengajuan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan not found")
		}
		return nil, err
	}

	return &pengajuan, nil
}

func (r *pengajuanIzinCutiRepository) Find(ctx context.Context, filter bson.M) ([]models.PengajuanIzinCuti, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pengajuans []models.PengajuanIzinCuti
	if err := cursor.All(ctx, &pengajuans); err != nil {
		return nil, err
	}

	return pengajuans, nil
}

func (r *pengajuanIzinCutiRepository) UpdateManagerHRDecision(ctx context.Context, id string, managerHRID primitive.ObjectID, statusManagerHR string, statusFinal string) (*models.PengajuanIzinCuti, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid pengajuan ID")
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "status_manager_hr": models.StatusPending}
	update := bson.M{
		"$set": bson.M{
			"manager_hr_id":     managerHRID,
			"status_manager_hr": statusManagerHR,
			"status_final":      statusFinal,
			"updated_at":        now,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.PengajuanIzinCuti
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("pengajuan sudah diproses")
		}
		return nil, err
	}

	return &updated, nil
}
