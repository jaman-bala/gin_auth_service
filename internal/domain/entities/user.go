package entities

import (
	"gold_portal/internal/errors"
	"gold_portal/internal/pkg/crypto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	FirstName  string    `gorm:"type:varchar(255);omitempty"`
	LastName   string    `gorm:"type:varchar(255);omitempty"`
	MiddleName string    `gorm:"type:varchar(255);omitempty"`
	Phone      string    `gorm:"unique;not null" validate:"required,e164"`
	Password   string    `gorm:"not null" validate:"required,min=8"`
	Role       Role      `gorm:"default:user" validate:"required"`
	Photo      string    `validate:"omitempty,url"`

	IsActive bool `gorm:"default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate хук GORM - автоматически hash пароль перед созданием
func (u *User) BeforeCreate(*gorm.DB) error {
	// Генерируем UUID если он не установлен
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Hash пароль если он не пустой
	if u.Password != "" {
		hashedPassword, err := crypto.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword
	}

	// Устанавливаем роль по умолчанию
	if u.Role == "" {
		u.Role = RoleUser
	}

	// Проверяем валидность роли
	if !u.Role.IsValid() {
		return errors.ErrInvalidUserRole
	}

	return nil
}

// BeforeUpdate хук GORM - hash пароль если он изменился
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Если пароль изменился, hash его
	if tx.Statement.Changed("Password") && u.Password != "" {
		hashedPassword, err := crypto.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword
	}

	// Проверяем валидность роли при изменении
	if tx.Statement.Changed("role") && !u.Role.IsValid() {
		return errors.ErrInvalidUserRole
	}

	return nil
}

// CheckPassword проверяет соответствие пароля
func (u *User) CheckPassword(password string) error {
	return crypto.CheckPassword(u.Password, password)
}
