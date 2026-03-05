// pkg/migration/migrations/002_create_positions.go
package migrations

import (
    "context"
    "time"

    "github.com/andikatampubolon10/hris-backend/pkg/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func CreatePositions() (int, string, string, func(*mongo.Database) error, func(*mongo.Database) error) {
    version := 2
    name := "create_positions"
    description := "Create positions collection and seed initial data"

    up := func(db *mongo.Database) error {
        ctx := context.Background()
        collection := db.Collection("positions")

        // Create indexes
        indexModels := []mongo.IndexModel{
            {
                Keys:    bson.D{{Key: "code", Value: 1}},
                Options: options.Index().SetUnique(true),
            },
            {
                Keys: bson.D{{Key: "department_id", Value: 1}},
            },
        }
        _, err := collection.Indexes().CreateMany(ctx, indexModels)
        if err != nil {
            return err
        }

        // Get departments for reference
        deptCollection := db.Collection("departments")
        deptCursor, err := deptCollection.Find(ctx, bson.M{})
        if err != nil {
            return err
        }
        defer deptCursor.Close(ctx)

        var departments []models.Department
        if err = deptCursor.All(ctx, &departments); err != nil {
            return err
        }

        deptMap := make(map[string]primitive.ObjectID)
        for _, dept := range departments {
            deptMap[dept.Code] = dept.ID
        }

        // Positions data
        positionsData := map[string][]struct {
            Code  string
            Name  string
            Level int
            Grade string
            Min   int
            Max   int
        }{
            "FO": {
                {"FO-01", "Front Office Manager", 5, "M1", 8000000, 12000000},
                {"FO-02", "Assistant Front Office Manager", 4, "A1", 6000000, 9000000},
                {"FO-03", "Front Desk Supervisor", 3, "S1", 5000000, 7000000},
                {"FO-04", "Receptionist", 2, "R1", 4000000, 6000000},
                {"FO-05", "Concierge", 2, "R1", 4000000, 6000000},
                {"FO-06", "Bell Boy", 1, "J1", 3500000, 4500000},
                {"FO-07", "Door Man", 1, "J1", 3500000, 4500000},
            },
            "HK": {
                {"HK-01", "Housekeeping Manager", 5, "M1", 7000000, 11000000},
                {"HK-02", "Assistant Housekeeping Manager", 4, "A1", 5500000, 8000000},
                {"HK-03", "Housekeeping Supervisor", 3, "S1", 4500000, 6500000},
                {"HK-04", "Room Attendant", 2, "R1", 3500000, 5000000},
            },
            "FB": {
                {"FB-01", "F&B Manager", 5, "M1", 9000000, 13000000},
                {"FB-02", "Restaurant Manager", 4, "A1", 6500000, 9500000},
                {"FB-03", "Chef de Cuisine", 4, "A1", 7000000, 10000000},
                {"FB-04", "Waiter/Waitress", 2, "R1", 3500000, 5000000},
            },
            "ENG": {
                {"ENG-01", "Chief Engineer", 5, "M1", 8000000, 12000000},
                {"ENG-02", "Technician", 2, "R1", 4000000, 6000000},
            },
            "HR": {
                {"HR-01", "HR Manager", 5, "M1", 8000000, 12000000},
                {"HR-02", "HR Staff", 2, "R1", 4000000, 6000000},
            },
        }

        // Build positions
        var positions []interface{}
        for deptCode, posList := range positionsData {
            deptID, exists := deptMap[deptCode]
            if !exists {
                continue
            }

            for _, p := range posList {
                positions = append(positions, models.Position{
                    ID:           primitive.NewObjectID(),
                    DepartmentID: deptID,
                    Code:         p.Code,
                    Name:         p.Name,
                    Level:        p.Level,
                    Grade:        p.Grade,
                    Description:  "Position: " + p.Name,
                    SalaryRange: models.SalaryRange{
                        Min:      p.Min,
                        Max:      p.Max,
                        Currency: "IDR",
                    },
                    IsActive:  true,
                    CreatedAt: time.Now(),
                    UpdatedAt: time.Now(),
                })
            }
        }

        if len(positions) > 0 {
            _, err = collection.InsertMany(ctx, positions)
            return err
        }

        return nil
    }

    down := func(db *mongo.Database) error {
        ctx := context.Background()
        return db.Collection("positions").Drop(ctx)
    }

    return version, name, description, up, down
}