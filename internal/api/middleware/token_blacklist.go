package middleware

import (
	"gold_portal/internal/domain/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenBlacklistMiddleware проверяет, не находится ли токен в черном списке
func TokenBlacklistMiddleware(tokenService services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Проверяем формат заголовка
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		// Проверяем, не в черном ли списке токен
		isBlacklisted, err := tokenService.IsTokenBlacklisted(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check token status",
			})
			c.Abort()
			return
		}

		if isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is blacklisted",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
