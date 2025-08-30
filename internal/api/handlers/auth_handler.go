package handlers

import (
	"gold_portal/internal/domain/dto"
	"gold_portal/internal/domain/entities"
	"gold_portal/internal/domain/services"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// UserRegister godoc
// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе с возможностью загрузки фото
// @Tags auth
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param first_name formData string true "Имя"
// @Param last_name formData string true "Фамилия"
// @Param middle_name formData string false "Отчество"
// @Param password formData string true "Пароль"
// @Param phone formData string true "Телефон"
// @Param photo formData file false "Фото профиля"
// @Success 201 {object} dto.UserResponseDTO
// @Router /api/v1/auth/web-register [post]
func (h *AuthHandler) UserRegister(c *gin.Context) {
	// Получаем данные формы
	var request dto.UserRequestDTO
	request.FirstName = c.PostForm("first_name")
	request.LastName = c.PostForm("last_name")
	request.MiddleName = c.PostForm("middle_name")
	request.Password = c.PostForm("password")
	request.Phone = c.PostForm("phone")

	var photoFile *multipart.FileHeader
	if file, err := c.FormFile("photo"); err == nil && file != nil {
		photoFile = file
	}

	ctx := c.Request.Context()
	user, err := h.authService.UserRegister(ctx, request, photoFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User success register",
		"user":    user,
	})
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе
// @Tags dashboard
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param first_name formData string true "Имя"
// @Param last_name formData string true "Фамилия"
// @Param middle_name formData string false "Отчество"
// @Param password formData string true "Пароль"
// @Param phone formData string true "Телефон"
// @Param role formData string false "Роль"
// @Param photo formData file false "Фото профиля"
// @Success 201 {object} dto.UserResponseDTO
// @Router /api/v1/dashboard/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request dto.UserDashboardDTO
	request.FirstName = c.PostForm("first_name")
	request.LastName = c.PostForm("last_name")
	request.MiddleName = c.PostForm("middle_name")
	request.Password = c.PostForm("password")
	request.Phone = c.PostForm("phone")
	request.Role = entities.Role(c.PostForm("role"))

	var photoFile *multipart.FileHeader
	if file, err := c.FormFile("photo"); err == nil && file != nil {
		photoFile = file
	}

	ctx := c.Request.Context()
	user, err := h.authService.Register(ctx, request, photoFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User success register",
		"user":    user,
	})
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентифицирует пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequestDTO true "Учетные данные"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx := c.Request.Context()
	tokenResponse, err := h.authService.Login(ctx, request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    tokenResponse.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.authService.GetAccessTokenExpiry().Seconds()),
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenResponse.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true в production с HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.authService.GetRefreshTokenExpiry().Seconds()),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenResponse.AccessToken,
		"message":      "Success authorization",
	})
}

// Logout godoc
// @Summary Выход из системы
// @Description Завершает сессию пользователя. На сервере токен не сохраняется, поэтому выход обрабатывается на клиенте (удаление токена).
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	token, err := c.Cookie("access_token")
	if err != nil || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Токен не предоставлен"})
		return
	}
	err = h.authService.Logout(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Success logout",
	})
}

// Refresh godoc
// @Summary Обновление access токена по refresh token
// @Description Получить новый access токен, отправив refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Тело запроса с refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Refresh token required"})
		return
	}

	ctx := c.Request.Context()
	tokenResponse, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenResponse.AccessToken,
		"message":      tokenResponse.Message,
	})
}

// UserMe
// @Summary Данные профиля
// @Description Данные авторизованного пользователя
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) UserMe(c *gin.Context) {
	id, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User is not authorized"})
	}
	ctx := c.Request.Context()
	profile, err := h.authService.UserMe(ctx, id.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}
