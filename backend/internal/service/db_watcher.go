// internal/service/db_watcher.go
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBWatcher listens to MongoDB Change Streams and broadcasts events in real-time.
type DBWatcher struct {
	db    *mongo.Database
	wsHub *WSHub
}

// NewDBWatcher creates a new database change stream watcher.
func NewDBWatcher(db *mongo.Database, wsHub *WSHub) *DBWatcher {
	return &DBWatcher{
		db:    db,
		wsHub: wsHub,
	}
}

// Start kicks off background change stream watchers for each collection.
func (w *DBWatcher) Start(ctx context.Context) {
	log.Println("[DB Watcher] Initializing real-time change stream watchers...")

	go w.watchCollection(ctx, "attendances", WSEventAttendanceUpdated, "user_id", "Kehadiran")
	go w.watchCollection(ctx, "leave_request", WSEventLeaveUpdated, "user_id", "Izin/Cuti")
	go w.watchCollection(ctx, "overtime_requests", WSEventOvertimeUpdated, "employees", "Lembur")
	go w.watchCollection(ctx, "assignments", WSEventAssignmentUpdated, "employees", "Penugasan")
}

func (w *DBWatcher) watchCollection(ctx context.Context, collectionName string, eventType WSEventType, userField string, friendlyName string) {
	collection := w.db.Collection(collectionName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[DB Watcher] Shutting down watcher for collection: %s", collectionName)
			return
		default:
			// Open a change stream on the collection
			pipeline := mongo.Pipeline{}
			streamOpts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
			stream, err := collection.Watch(ctx, pipeline, streamOpts)
			if err != nil {
				log.Printf("[DB Watcher] Error starting change stream for %s: %v. Retrying in 5 seconds...", collectionName, err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("[DB Watcher] Active and listening for changes on: %s", collectionName)

			for stream.Next(ctx) {
				var changeDoc bson.M
				if err := stream.Decode(&changeDoc); err != nil {
					log.Printf("[DB Watcher] Decode error on stream %s: %v", collectionName, err)
					continue
				}

				operationType, _ := changeDoc["operationType"].(string)
				log.Printf("[DB Watcher] Detected mutation (%s) in collection: %s", operationType, collectionName)

				// Extract full document if present
				fullDoc, ok := changeDoc["fullDocument"].(bson.M)

				// 1. Silent reload event for ALL active connections
				// This forces immediate, silent refresh in the Flutter UI for any user viewing the page.
				w.wsHub.BroadcastToAll(eventType, map[string]any{
					"action":    operationType,
					"silent":    true,
					"timestamp": time.Now().Unix(),
				})

				// Also trigger stats refresh in the dashboard
				w.wsHub.BroadcastToAll(WSEventStatsUpdated, map[string]any{
					"reason": "db_change",
				})

				// 2. Target specific users to show Premium In-App Notification Banners (red dots & banners)
				if ok && fullDoc != nil {
					// Handle collections where the user is in a single field (user_id)
					if userField == "user_id" {
						if userIDVal, exists := fullDoc["user_id"]; exists {
							var userIDStr string
							if oid, isOid := userIDVal.(primitive.ObjectID); isOid {
								userIDStr = oid.Hex()
							} else if s, isStr := userIDVal.(string); isStr {
								userIDStr = s
							}

							if userIDStr != "" {
								actionMsg := "diperbarui"
								if operationType == "insert" {
									actionMsg = "dibuat"
								} else if operationType == "delete" {
									actionMsg = "dihapus"
								}

								w.wsHub.BroadcastToUser(userIDStr, eventType, map[string]any{
									"action":  operationType,
									"message": fmt.Sprintf("Data %s Anda telah %s", friendlyName, actionMsg),
								})
							}
						}
					} else if userField == "employees" {
						// Handle collections where users are listed in an array of employees (overtime, assignments)
						if employeesVal, exists := fullDoc["employees"]; exists {
							if empList, ok := employeesVal.(primitive.A); ok {
								for _, empItem := range empList {
									if empMap, isMap := empItem.(bson.M); isMap {
										if userIDVal, exists := empMap["user_id"]; exists {
											var userIDStr string
											if oid, isOid := userIDVal.(primitive.ObjectID); isOid {
												userIDStr = oid.Hex()
											} else if s, isStr := userIDVal.(string); isStr {
												userIDStr = s
											}

											if userIDStr != "" {
												actionMsg := "diperbarui"
												if operationType == "insert" {
													actionMsg = "diberikan untuk Anda"
												}

												w.wsHub.BroadcastToUser(userIDStr, eventType, map[string]any{
													"action":  operationType,
													"message": fmt.Sprintf("Data %s baru telah %s", friendlyName, actionMsg),
												})
											}
										}
									}
								}
							}
						}
					}
				}
			}

			stream.Close(ctx)
		}
	}
}
