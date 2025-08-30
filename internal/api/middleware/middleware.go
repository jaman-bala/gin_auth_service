package middleware

import (
	"fmt"
	"gold_portal/internal/domain/services"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AuthorizationHeaderKey = "Authorization"
	BearerSchema           = "Bearer "
	AccessTokenCookieName  = "access_token"
)

// AuthMiddleware middleware для проверки JWT токена
func AuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Попробовать получить токен из cookie
		if cookieToken, err := c.Cookie(AccessTokenCookieName); err == nil && cookieToken != "" {
			tokenString = cookieToken
		} else {
			// Если в cookie нет — пробуем из заголовка Authorization
			authHeader := c.GetHeader(AuthorizationHeaderKey)
			if strings.HasPrefix(authHeader, BearerSchema) {
				tokenString = strings.TrimPrefix(authHeader, BearerSchema)
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Отсутствует токен авторизации",
				"code":  "AUTH_TOKEN_MISSING",
			})
			c.Abort()
			return
		}

		// Валидация токена
		token, err := authService.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Недействительный токен",
				"code":    "AUTH_TOKEN_INVALID",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Получение пользователя из токена
		userDTO, err := authService.GetUserFromToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Ошибка получения пользователя из токена",
				"code":    "AUTH_USER_NOT_FOUND",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Добавление пользователя в контекст запроса
		c.Set("id", userDTO.ID)
		c.Set("user", userDTO)
		c.Set("role", userDTO.Role)

		c.Next()
	}
}

func AuditMiddleware(auditService services.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Получаем IP адрес клиента
		clientIP := c.ClientIP()

		// Получаем User-Agent
		userAgent := c.GetHeader("User-Agent")

		c.Next() // выполняем основной обработчик

		status := c.Writer.Status()

		// ПОСЛЕ c.Next() получаем данные пользователя из контекста
		var userID = uuid.Nil
		var userName = "Гость"

		if uid, exists := c.Get("id"); exists {
			if uuidVal, ok := uid.(uuid.UUID); ok {
				userID = uuidVal
				userName = "Авторизованный пользователь"
			}
		}

		// Определяем тип объекта из URL
		var entityType = "Unknown"
		var entityID = uuid.Nil

		// Извлекаем ID из URL параметров
		if idParam := c.Param("id"); idParam != "" {
			if parsedID, err := uuid.Parse(idParam); err == nil {
				entityID = parsedID
			}
		}

		// Определяем тип объекта по пути
		if strings.Contains(path, "/products") {
			entityType = "Product"
		} else if strings.Contains(path, "/users") || strings.Contains(path, "/dashboard") {
			entityType = "User"
		} else if strings.Contains(path, "/orders") {
			entityType = "Order"
		} else if strings.Contains(path, "/auth") {
			entityType = "Auth"
		} else if strings.Contains(path, "/audit") {
			entityType = "Audit"
		}

		// Создаем описание действия
		var actionDescription string
		switch method {
		case "GET":
			actionDescription = "Просмотр"
		case "POST":
			actionDescription = "Создание"
		case "PUT", "PATCH":
			actionDescription = "Обновление"
		case "DELETE":
			actionDescription = "Удаление"
		default:
			actionDescription = method
		}

		// Формируем данные для логирования
		logData := fmt.Sprintf("Действие: %s %s | Пользователь: %s", actionDescription, entityType, userName)

		// Логируем действие
		_ = auditService.Log(userID, entityID, method, path, status, clientIP, userAgent, logData)
	}
}

// AdminRoleMiddleware проверяет, что пользователь имеет роль "admin"
func AdminRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Роль пользователя не определена",
				"code":  "AUTH_ROLE_MISSING",
			})
			return
		}

		roleStr := fmt.Sprintf("%v", roleValue)

		if roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Недостаточно прав для выполнения операции",
				"code":          "AUTH_INSUFFICIENT_PRIVILEGES",
				"required_role": "admin",
				"user_role":     roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func SuperUserRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Роль пользователя не определена",
				"code":  "AUTH_ROLE_MISSING",
			})
			return
		}

		roleStr := fmt.Sprintf("%v", roleValue)

		if roleStr != "superuser" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Недостаточно прав для выполнения операции",
				"code":          "AUTH_INSUFFICIENT_PRIVILEGES",
				"required_role": "superuser",
				"user_role":     roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ManagerRoleMiddleware проверяет, что пользователь имеет роль "manager" или выше
func ManagerRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Роль пользователя не определена",
				"code":  "AUTH_ROLE_MISSING",
			})
			return
		}

		roleStr := fmt.Sprintf("%v", roleValue)

		// Проверяем, что роль manager или выше (admin, superuser)
		if roleStr != "manager" && roleStr != "admin" && roleStr != "superuser" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Недостаточно прав для выполнения операции",
				"code":          "AUTH_INSUFFICIENT_PRIVILEGES",
				"required_role": "manager",
				"user_role":     roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRoleMiddleware проверяет, что пользователь имеет одну из указанных ролей
func RequireRoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Роль пользователя не определена",
				"code":  "AUTH_ROLE_MISSING",
			})
			return
		}

		userRole := fmt.Sprintf("%v", roleValue)

		// Проверяем, есть ли роль пользователя в списке разрешенных ролей
		for _, requiredRole := range roles {
			if userRole == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":          "Недостаточно прав для выполнения операции",
			"code":           "AUTH_INSUFFICIENT_PRIVILEGES",
			"required_roles": roles,
			"user_role":      userRole,
		})
		c.Abort()
	}
}

// RequireRoleLevelMiddleware проверяет, что пользователь имеет роль с уровнем не ниже указанного
func RequireRoleLevelMiddleware(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Роль пользователя не определена",
				"code":  "AUTH_ROLE_MISSING",
			})
			return
		}

		roleStr := fmt.Sprintf("%v", roleValue)

		// Определяем уровень роли пользователя
		var userLevel int
		switch roleStr {
		case "superuser":
			userLevel = 4
		case "admin":
			userLevel = 3
		case "manager":
			userLevel = 2
		case "user":
			userLevel = 1
		default:
			userLevel = 0
		}

		if userLevel < minLevel {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Недостаточно прав для выполнения операции",
				"code":           "AUTH_INSUFFICIENT_PRIVILEGES",
				"required_level": minLevel,
				"user_level":     userLevel,
				"user_role":      roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware базовый middleware для ограничения запросов
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// Простая реализация rate limiting с использованием map
	// В продакшене лучше использовать Redis или специализированные библиотеки
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Очистка старых записей
		if times, exists := clients[clientIP]; exists {
			var validTimes []time.Time
			for _, t := range times {
				if now.Sub(t) < time.Minute {
					validTimes = append(validTimes, t)
				}
			}
			clients[clientIP] = validTimes
		}

		// Проверка лимита
		if len(clients[clientIP]) >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Превышен лимит запросов",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": "60 seconds",
			})
			c.Abort()
			return
		}

		// Добавление текущего запроса
		clients[clientIP] = append(clients[clientIP], now)

		c.Next()
	}
}
