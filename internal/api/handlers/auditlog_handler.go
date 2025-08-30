package handlers

import (
	"gold_portal/internal/domain/dto"
	"gold_portal/internal/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService services.AuditService
}

func NewAuditHandler(auditService services.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// GetAllLogs godoc
// @Summary Получить все действия пользователей
// @Description Возвращает список действий пользователей
// @Tags audit
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.AuditLogResponse
// @Router /api/v1/audit [get]
func (h *AuditHandler) GetAllLogs(c *gin.Context) {
	logs, err := h.auditService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем entities в DTO
	var responseLogs []dto.AuditLogResponse
	for _, log := range logs {
		responseLog := dto.AuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    log.Action,
			Entity:    log.Entity,
			EntityID:  log.EntityID,
			Data:      log.Data,
			Status:    log.Status,
			ClientIP:  log.ClientIP,
			UserAgent: log.UserAgent,
			CreatedAt: log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		responseLogs = append(responseLogs, responseLog)
	}

	c.JSON(http.StatusOK, responseLogs)
}
