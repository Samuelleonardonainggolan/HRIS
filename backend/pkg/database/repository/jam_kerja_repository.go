// pkg/database/repository/jam_kerja_repository.go
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

type JamKerjaRepository interface {
	Create(ctx context.Context, jamKerja *models.JamKerja) error
	FindByID(ctx context.Context, id string) (*models.JamKerja, error)
	FindByUserID(ctx context.Context, userID string) (*models.JamKerja, error)
	FindAll(ctx context.Context) ([]models.JamKerja, error)
	UpsertByUserID(ctx context.Context, userID string, req *models.JamKerja) error
	Update(ctx context.Context, id string, req *models.JamKerja) error
	Delete(ctx context.Context, id string) error
	GetAllUserIDs(ctx context.Context) ([]primitive.ObjectID, error)
}

type jamKerjaRepository struct {
	collection *mongo.Collection
}

func NewJamKerjaRepository(db *mongo.Database) JamKerjaRepository {
	return &jamKerjaRepository{
		collection: db.Collection("jam_kerja"),
	}
}

func (r *jamKerjaRepository) Create(ctx context.Context, jamKerja *models.JamKerja) error {
	jamKerja.ID = primitive.NewObjectID()
	jamKerja.CreatedAt = time.Now()
	jamKerja.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, jamKerja)
	return err
}

func (r *jamKerjaRepository) FindByID(ctx context.Context, id string) (*models.JamKerja, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid jam kerja ID")
	}

	var jk models.JamKerja
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&jk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("jam kerja not found")
		}
		return nil, err
	}

	return &jk, nil
}

func (r *jamKerjaRepository) FindByUserID(ctx context.Context, userID string) (*models.JamKerja, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var jk models.JamKerja
	err = r.collection.FindOne(ctx, bson.M{"user_id": oid}).Decode(&jk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // boleh nil kalau belum ada schedule
		}
		return nil, err
	}
	return &jk, nil
}

func (r *jamKerjaRepository) FindAll(ctx context.Context) ([]models.JamKerja, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cur, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []models.JamKerja
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *jamKerjaRepository) UpsertByUserID(ctx context.Context, userID string, req *models.JamKerja) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"user_id":        oid,
			"hari_kerja":     req.HariKerja,
			"waktu_mulai":    req.WaktuMulai,
			"waktu_selesai":  req.WaktuSelesai,
			"aktif":          req.Aktif,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": oid}, update, options.Update().SetUpsert(true))
	return err
}

func (r *jamKerjaRepository) Update(ctx context.Context, id string, req *models.JamKerja) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid jam kerja ID")
	}

	update := bson.M{
		"$set": bson.M{
			"hari_kerja":    req.HariKerja,
			"waktu_mulai":   req.WaktuMulai,
			"waktu_selesai": req.WaktuSelesai,
			"aktif":         req.Aktif,
			"updated_at":    time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("jam kerja not found")
	}
	return nil
}

func (r *jamKerjaRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid jam kerja ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("jam kerja not found")
	}
	return nil
}

func (r *jamKerjaRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, errors.New("invalid user ID")
	}

	err = r.collection.FindOne(ctx, bson.M{"user_id": oid}, options.FindOne().SetProjection(bson.M{"_id": 1})).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *jamKerjaRepository) GetAllUserIDs(ctx context.Context) ([]primitive.ObjectID, error) {
	cur, err := r.collection.Find(
		ctx,
		bson.M{},
		options.Find().SetProjection(bson.M{"user_id": 1}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	type row struct {
		UserID primitive.ObjectID `bson:"user_id"`
	}

	var ids []primitive.ObjectID
	for cur.Next(ctx) {
		var x row
		if err := cur.Decode(&x); err != nil {
			return nil, err
		}
		if x.UserID != primitive.NilObjectID {
			ids = append(ids, x.UserID)
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
