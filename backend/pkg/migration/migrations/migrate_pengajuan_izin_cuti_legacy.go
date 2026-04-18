package migrations

import (
	"context"
	"fmt"

	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MigratePengajuanIzinCutiLegacy() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
	version := 12
	name := "migrate_pengajuan_izin_cuti_legacy"
	description := "Copy legacy pengajuan_izin_cuti documents into leave_request with proper kepala_departemen_id"

	up := func(db *mongo.Database) error {
		ctx := context.Background()

		oldCol := db.Collection("pengajuan_izin_cuti")
		newCol := db.Collection("leave_request")
		usersCol := db.Collection("users")
		deptsCol := db.Collection("departments")

		var managerHR struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		_ = usersCol.FindOne(ctx, bson.M{"role": models.RoleManagerHR}, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&managerHR)

		cur, err := oldCol.Find(ctx, bson.M{})
		if err != nil {
			return err
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var p models.LeaveRequest
			if err := cur.Decode(&p); err != nil {
				return err
			}

			var requester models.User
			if err := usersCol.FindOne(ctx, bson.M{"_id": p.UserID}).Decode(&requester); err == nil {
				var dept models.Department
				if err := deptsCol.FindOne(ctx, bson.M{"_id": requester.DepartmentID}).Decode(&dept); err == nil {
					if !dept.ManagerID.IsZero() {
						p.KepalaDepartemenID = dept.ManagerID
					}
				}
			}

			if p.ManagerHRID.IsZero() && !managerHR.ID.IsZero() {
				p.ManagerHRID = managerHR.ID
			}
			if p.KepalaDepartemenID.IsZero() && !managerHR.ID.IsZero() {
				p.KepalaDepartemenID = managerHR.ID
			}

			_, err := newCol.ReplaceOne(
				ctx,
				bson.M{"_id": p.ID},
				p,
				options.Replace().SetUpsert(true),
			)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					continue
				}
				return fmt.Errorf("failed to upsert leave_request %s: %w", p.ID.Hex(), err)
			}
		}
		if err := cur.Err(); err != nil {
			return err
		}

		return nil
	}

	down := func(db *mongo.Database) error {
		return nil
	}

	return version, name, description, up, down
}

