// internal/service/notification_service.go
package service

import (
	"context"
	"errors"

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
}

type notificationService struct {
	repo  repository.NotificationRepository
	wsHub *WSHub
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{
		repo: repo,
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
