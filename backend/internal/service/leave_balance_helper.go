package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultAnnualLeaveQuota = 12

func requestTypeConsumesQuota(ctx context.Context, db *mongo.Database, requestTypeID primitive.ObjectID) (bool, error) {
	var requestType models.RequestType
	err := db.Collection("request_type").FindOne(ctx, bson.M{"_id": requestTypeID}).Decode(&requestType)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("tipe pengajuan tidak ditemukan")
		}
		return false, err
	}

	return categoryConsumesQuota(requestType.CategoryName) || requestType.QuotaDeduction || requestType.PotongKuota, nil
}

func syncLeaveBalanceForYear(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, year int) (*models.LeaveBalance, error) {
	balance, err := ensureLeaveBalanceForYear(ctx, db, userID, year)
	if err != nil {
		return nil, err
	}

	usedQuota, err := calculateUsedQuotaForYear(ctx, db, userID, year)
	if err != nil {
		return nil, err
	}

	totalQuota := balance.TotalKuota
	if totalQuota <= 0 {
		totalQuota = defaultAnnualLeaveQuota
	}

	remainingQuota := totalQuota - usedQuota
	if remainingQuota < 0 {
		remainingQuota = 0
	}

	now := time.Now()
	_, err = db.Collection("leave_balance").UpdateOne(
		ctx,
		bson.M{"_id": balance.ID},
		bson.M{"$set": bson.M{
			"total_kuota":     totalQuota,
			"used_kuota":      usedQuota,
			"remaining_kuota": remainingQuota,
			"updated_at":      now,
		}},
	)
	if err != nil {
		return nil, err
	}

	var updated models.LeaveBalance
	if err := db.Collection("leave_balance").FindOne(ctx, bson.M{"_id": balance.ID}).Decode(&updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

func ensureLeaveBalanceForYear(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, year int) (*models.LeaveBalance, error) {
	var balance models.LeaveBalance
	err := db.Collection("leave_balance").FindOne(ctx, bson.M{"user_id": userID, "year": year}).Decode(&balance)
	if err == nil {
		return &balance, nil
	}
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	now := time.Now()
	balance = models.LeaveBalance{
		ID:             primitive.NewObjectID(),
		UserID:         userID,
		Year:           year,
		TotalKuota:     defaultAnnualLeaveQuota,
		UsedKuota:      0,
		RemainingKuota: defaultAnnualLeaveQuota,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if _, err := db.Collection("leave_balance").InsertOne(ctx, balance); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ensureLeaveBalanceForYear(ctx, db, userID, year)
		}
		return nil, err
	}

	return &balance, nil
}

func calculateUsedQuotaForYear(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, year int) (int, error) {
	quotaTypeIDs, err := getQuotaDeductingRequestTypeIDs(ctx, db)
	if err != nil {
		return 0, err
	}
	if len(quotaTypeIDs) == 0 {
		return 0, nil
	}

	startOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	startOfNextYear := startOfYear.AddDate(1, 0, 0)

	cursor, err := db.Collection("leave_request").Find(ctx, bson.M{
		"user_id":         userID,
		"final_status":    models.StatusApproved,
		"request_type_id": bson.M{"$in": quotaTypeIDs},
		"start_date":      bson.M{"$gte": startOfYear, "$lt": startOfNextYear},
	})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	type usageRow struct {
		DaysTotal int `bson:"days_total"`
	}

	var rows []usageRow
	if err := cursor.All(ctx, &rows); err != nil {
		return 0, err
	}

	usedQuota := 0
	for _, row := range rows {
		usedQuota += row.DaysTotal
	}

	return usedQuota, nil
}

func getQuotaDeductingRequestTypeIDs(ctx context.Context, db *mongo.Database) ([]primitive.ObjectID, error) {
	cursor, err := db.Collection("request_type").Find(ctx, bson.M{
		"$or": []bson.M{
			{"category_name": bson.M{"$in": []string{"Izin", "Cuti"}}},
			{"quota_deduction": true},
			{"potong_kuota": true},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type requestTypeRow struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	var rows []requestTypeRow
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	ids := make([]primitive.ObjectID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	return ids, nil
}

func categoryConsumesQuota(categoryName string) bool {
	switch strings.ToLower(strings.TrimSpace(categoryName)) {
	case "izin", "cuti":
		return true
	default:
		return false
	}
}
