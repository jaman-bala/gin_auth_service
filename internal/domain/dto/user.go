package dto

import (
	"gold_portal/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID         uuid.UUID     `json:"id"`
	FirstName  string        `json:"first_name"`
	LastName   string        `json:"last_name"`
	MiddleName string        `json:"middle_name"`
	Phone      string        `json:"phone" validate:"required,e164" example:"+996500500500"`
	Role       entities.Role `json:"role" default:"user"`
	Photo      string        `json:"photo" validate:"omitempty,url"`
	IsActive   bool          `json:"is_active"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	DeletedAt  *time.Time    `json:"deleted_at"`
}

type UserRequestDTO struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Password   string `json:"password" validate:"required,min=8" example:"Password123"`
	Phone      string `json:"phone" validate:"required,e164" example:"+996500500500"`
	Photo      string `json:"photo" validate:"omitempty,url"`
}
type UserDashboardDTO struct {
	FirstName  string        `json:"first_name"`
	LastName   string        `json:"last_name"`
	MiddleName string        `json:"middle_name"`
	Password   string        `json:"password" validate:"required,min=8" example:"Password123"`
	Phone      string        `json:"phone" validate:"required,e164" example:"+996500500500"`
	Role       entities.Role `json:"role" default:"user"`
	Photo      string        `json:"photo" validate:"omitempty,url"`
}

type UserResponseDTO struct {
	ID         uuid.UUID     `json:"id"`
	FirstName  string        `json:"first_name"`
	LastName   string        `json:"last_name"`
	MiddleName string        `json:"middle_name"`
	Phone      string        `json:"phone"`
	Role       entities.Role `json:"role"`
	Photo      string        `json:"photo"`
	IsActive   bool          `json:"is_active"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	DeletedAt  *time.Time    `json:"deleted_at"`
}

type UserUpdateDTO struct {
	FirstName  *string        `json:"first_name"`
	LastName   *string        `json:"last_name"`
	MiddleName *string        `json:"middle_name"`
	Phone      *string        `json:"phone" validate:"required,e164" example:"+996500500500"`
	Role       *entities.Role `json:"role" default:"user"`
	Photo      *string        `json:"photo"`
	IsActive   *bool          `json:"is_active"`
}

type LoginRequestDTO struct {
	Phone    string `json:"phone" validate:"required,e164" example:"+996500500500"`
	Password string `json:"password" validate:"required,min=8" example:"Password123"`
}

type LoginResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Message      string `json:"message"`
}

type TokenResponseDTO struct {
	AccessToken string          `json:"access_token"`
	User        UserResponseDTO `json:"user"`
	ExpiresAt   time.Time       `json:"expires_at"`
	Message     string          `json:"message,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token"`
}

type ChangePasswordDashboardDTO struct {
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type ChangePasswordDTO struct {
	OldPassword     string `json:"old_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

func (dto *UserResponseDTO) FromModel(user *entities.User) {
	dto.ID = user.ID
	dto.FirstName = user.FirstName
	dto.LastName = user.LastName
	dto.MiddleName = user.MiddleName
	dto.Phone = user.Phone
	dto.Role = user.Role
	dto.Photo = user.Photo
	dto.IsActive = user.IsActive
	dto.CreatedAt = user.CreatedAt
	dto.UpdatedAt = user.UpdatedAt
	if user.DeletedAt.Valid {
		t := user.DeletedAt.Time
		dto.DeletedAt = &t
	} else {
		dto.DeletedAt = nil
	}
}

func (dto *UserResponseDTO) FromModelUser(user *entities.User) {
	dto.ID = user.ID
	dto.FirstName = user.FirstName
	dto.LastName = user.LastName
	dto.MiddleName = user.MiddleName
	dto.Phone = user.Phone
	dto.Photo = user.Photo
	dto.IsActive = user.IsActive
	dto.CreatedAt = user.CreatedAt
	dto.UpdatedAt = user.UpdatedAt
	if user.DeletedAt.Valid {
		t := user.DeletedAt.Time
		dto.DeletedAt = &t
	} else {
		dto.DeletedAt = nil
	}
}

func (dto *UserDashboardDTO) ToModel(hashedPassword string) *entities.User {
	user := &entities.User{
		FirstName:  dto.FirstName,
		LastName:   dto.LastName,
		MiddleName: dto.MiddleName,
		Password:   dto.Password,
		Phone:      dto.Phone,
		Role:       dto.Role,
		Photo:      dto.Photo,
		IsActive:   true,
	}
	if user.Role == "" {
		user.Role = entities.RoleUser
	}

	return user
}

func (dto *UserRequestDTO) ToModelUser(hashedPassword string) *entities.User {
	user := &entities.User{
		FirstName:  dto.FirstName,
		LastName:   dto.LastName,
		MiddleName: dto.MiddleName,
		Password:   dto.Password,
		Phone:      dto.Phone,
		Photo:      dto.Photo,
		IsActive:   true,
	}
	if user.Role == "" {
		user.Role = entities.RoleUser
	}

	return user
}

func (dto *UserUpdateDTO) ApplyToModel(user *entities.User) {
	if dto.FirstName != nil {
		user.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		user.LastName = *dto.LastName
	}
	if dto.MiddleName != nil {
		user.MiddleName = *dto.MiddleName
	}
	if dto.Phone != nil {
		user.Phone = *dto.Phone
	}
	if dto.Role != nil {
		user.Role = *dto.Role
	}
	if dto.Photo != nil {
		user.Photo = *dto.Photo
	}
	if dto.IsActive != nil {
		user.IsActive = *dto.IsActive
	}
}
