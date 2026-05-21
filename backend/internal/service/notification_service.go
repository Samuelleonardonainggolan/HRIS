// internal/service/notification_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error)
	GetNotificationsByUserID(ctx context.Context, userID string, limit int) ([]models.NotificationResponse, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	SetWSHub(hub *WSHub)
	RegisterDeviceToken(ctx context.Context, userID string, token string) error
	UnregisterDeviceToken(ctx context.Context, token string) error
}

type notificationService struct {
	repo            repository.NotificationRepository
	deviceTokenRepo repository.DeviceTokenRepository
	wsHub           *WSHub
}

func NewNotificationService(repo repository.NotificationRepository, deviceRepo repository.DeviceTokenRepository) NotificationService {
	return &notificationService{
		repo:            repo,
		deviceTokenRepo: deviceRepo,
	}
}

func (s *notificationService) SetWSHub(hub *WSHub) {
	s.wsHub = hub
}

func (s *notificationService) CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error) {
	uid, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	n := &models.Notification{
		UserID:  uid,
		Title:   req.Title,
		Message: req.Message,
		Type:    req.Type,
	}

	if req.SenderID != "" {
		senderID, err := primitive.ObjectIDFromHex(req.SenderID)
		if err == nil {
			n.SenderID = senderID
		}
	}

	if req.ReferenceID != "" {
		refID, err := primitive.ObjectIDFromHex(req.ReferenceID)
		if err == nil {
			n.ReferenceID = refID
		}
	}

	err = s.repo.Create(ctx, n)
	if err != nil {
		return nil, err
	}

	resp := n.ToResponse()

	// Broadcast SSE event to target user
	if s.wsHub != nil {
		s.wsHub.BroadcastToUser(req.UserID, WSEventNotificationCreated, map[string]any{
			"id":           resp.ID,
			"sender_id":    resp.SenderID,
			"title":        resp.Title,
			"message":      resp.Message,
			"type":         resp.Type,
			"reference_id": resp.ReferenceID,
			"created_at":   resp.CreatedAt,
			"is_read":      resp.IsRead,
		})
	}

	// Send push via FCM if server key provided and device tokens available
	if s.deviceTokenRepo != nil {
		tokens, _ := s.deviceTokenRepo.FindByUserID(ctx, req.UserID)
		if len(tokens) > 0 {
			go sendFCM(tokens, req.Title, req.Message, map[string]string{"type": req.Type, "reference_id": req.ReferenceID})
		}
	}

	return &resp, nil
}

func (s *notificationService) GetNotificationsByUserID(ctx context.Context, userID string, limit int) ([]models.NotificationResponse, error) {
	list, err := s.repo.FindByUserID(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	resps := make([]models.NotificationResponse, len(list))
	for i, n := range list {
		resps[i] = n.ToResponse()
	}
	return resps, nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id string) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) RegisterDeviceToken(ctx context.Context, userID string, token string) error {
	if s.deviceTokenRepo == nil {
		return errors.New("device token repository not configured")
	}
	return s.deviceTokenRepo.Save(ctx, userID, token)
}

func (s *notificationService) UnregisterDeviceToken(ctx context.Context, token string) error {
	if s.deviceTokenRepo == nil {
		return errors.New("device token repository not configured")
	}
	return s.deviceTokenRepo.Delete(ctx, token)
}

// sendFCM sends a simple notification via legacy FCM HTTP API using server key
func sendFCM(tokens []string, title, body string, data map[string]string) {
	serverKey := os.Getenv("FCM_SERVER_KEY")
	if serverKey == "" {
		return
	}

	payload := map[string]any{}
	if len(tokens) == 1 {
		payload["to"] = tokens[0]
	} else {
		payload["registration_ids"] = tokens
	}
	payload["notification"] = map[string]string{"title": title, "body": body}
	if len(data) > 0 {
		payload["data"] = data
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "key="+serverKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err == nil && resp.Body != nil {
		resp.Body.Close()
	}
	_ = err
}
