package entities

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Action    string
	Entity    string
	EntityID  uuid.UUID `gorm:"type:uuid"`
	ClientIP  string
	UserAgent string
	Data      string
	Status    int
	CreatedAt time.Time
}
