package dto

import "github.com/google/uuid"

type AuditLogResponse struct {
	ID        uint      `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  uuid.UUID `json:"entity_id"`
	Data      string    `json:"description"` // Описание действия на русском
	Status    int       `json:"status"`
	ClientIP  string    `json:"client_ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt string    `json:"created_at"`
}
