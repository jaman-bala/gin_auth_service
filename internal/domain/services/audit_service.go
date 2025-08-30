package services

import (
	"gold_portal/internal/domain/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditService interface {
	Log(userID, entityID uuid.UUID, action, entity string, status int, clientIP, userAgent, data string) error
	GetAll() ([]entities.AuditLog, error)
	GetByUserID(userID uuid.UUID) ([]entities.AuditLog, error)
	GetByEntity(entity string) ([]entities.AuditLog, error)
}

type auditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) AuditService {
	return &auditService{db: db}
}

func (s *auditService) Log(userID, entityID uuid.UUID, action, entity string, status int, clientIP, userAgent, data string) error {
	log := entities.AuditLog{
		UserID:    userID,
		EntityID:  entityID,
		Action:    action,
		Entity:    entity,
		Status:    status,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Data:      data,
		CreatedAt: time.Now(),
	}
	return s.db.Create(&log).Error
}

func (s *auditService) GetAll() ([]entities.AuditLog, error) {
	var logs []entities.AuditLog
	if err := s.db.Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *auditService) GetByUserID(userID uuid.UUID) ([]entities.AuditLog, error) {
	var logs []entities.AuditLog
	if err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *auditService) GetByEntity(entity string) ([]entities.AuditLog, error) {
	var logs []entities.AuditLog
	if err := s.db.Where("entity = ?", entity).Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
