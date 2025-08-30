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

type UserHandler struct {
	userService services.UsersService
}

func NewUserHandler(userService services.UsersService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetAll godoc
// @Summary Получение всех пользователей
// @Description Возвращает список всех зарегистрированных пользователей (только для авторизованных пользователей)
// @Tags dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.UserResponseDTO "Список пользователей"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/dashboard [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	users, err := h.userService.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetByID godoc
// @Summary Получение пользователя по ID
// @Description Возвращает информацию о пользователе по указанному ID
// @Tags dashboard
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Router /api/v1/dashboard/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	userParam := c.Param("id")
	userID, err := uuid.Parse(userParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Некорректный ID пользователя"})
		return
	}
	ctx := c.Request.Context()
	user, err := h.userService.UserID(ctx, userID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": "Пользователь не найден"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetByPhone godoc
// @Summary Получить пользователя по номеру телефона
// @Description Возвращает информацию о пользователе по номеру телефона
// @Tags dashboard
// @Security BearerAuth
// @Param phone path string true "Номер телефона пользователя" Example("+996500500500")
// @Produce json
// @Success 200 {object} dto.UserResponseDTO "Информация о пользователе"
// @Router  /api/v1/dashboard/phone/{phone} [get]
func (h *UserHandler) GetByPhone(c *gin.Context) {
	phone := c.Param("phone")
	ctx := c.Request.Context()
	user, err := h.userService.GetByPhone(ctx, phone)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": "Пользователь не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}

// Patch godoc
// @Summary Обновление пользователя
// @Description Обновляет информацию о пользователе с возможностью загрузки фото. Можно передавать только те поля, которые нужно изменить (PATCH).
// @Tags dashboard
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID пользователя"
// @Param first_name formData string false "Имя"
// @Param last_name formData string false "Фамилия"
// @Param middle_name formData string false "Отчество"
// @Param phone formData string false "Телефон"
// @Param role formData string false "Роль"
// @Param is_active formData bool false "Активен"
// @Param photo formData file false "Фото профиля"
// @Router /api/v1/dashboard/patch/{id} [patch]
func (h *UserHandler) Patch(c *gin.Context) {
	userParam := c.Param("id")
	userID, err := uuid.Parse(userParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Некорректный ID пользователя"})
		return
	}

	// Получаем данные формы
	var request dto.UserUpdateDTO

	if firstName := c.PostForm("first_name"); firstName != "" {
		request.FirstName = &firstName
	}
	if lastName := c.PostForm("last_name"); lastName != "" {
		request.LastName = &lastName
	}
	if middleName := c.PostForm("middle_name"); middleName != "" {
		request.MiddleName = &middleName
	}
	if phone := c.PostForm("phone"); phone != "" {
		request.Phone = &phone
	}
	if role := c.PostForm("role"); role != "" {
		userRole := entities.Role(role)
		request.Role = &userRole
	}
	if isActiveStr := c.PostForm("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		request.IsActive = &isActive
	}

	// Получаем файл фото
	var photoFile *multipart.FileHeader
	if file, err := c.FormFile("photo"); err == nil && file != nil {
		photoFile = file
	}

	ctx := c.Request.Context()
	user, err := h.userService.Patch(ctx, userID, request, photoFile)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": "Пользователь не найден"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary Удаление пользователя
// @Description Удаляет пользователя по указанному ID
// @Tags dashboard
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Success 204 "Пользователь успешно удален"
// @Router /api/v1/dashboard/delete/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	userParam := c.Param("id")
	userID, err := uuid.Parse(userParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Некорректный ID пользователя"})
		return
	}
	ctx := c.Request.Context()
	err = h.userService.Delete(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удалён"})
}
