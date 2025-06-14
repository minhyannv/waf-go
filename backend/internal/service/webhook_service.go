package service

import (
	"waf-go/internal/models"

	"gorm.io/gorm"
)

type WebhookService struct {
	db *gorm.DB
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

func (s *WebhookService) GetWebhookList(tenantID uint) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	query := s.db.Model(&models.Webhook{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.Find(&webhooks).Error
	return webhooks, err
}
