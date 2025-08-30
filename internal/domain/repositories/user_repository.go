package repositories

import (
	"context"
	stdErrors "errors"
	"gold_portal/internal/domain/entities"
	"gold_portal/internal/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	Get(ctx context.Context) ([]*entities.User, error)
	GetID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	Patch(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	FindByPhone(ctx context.Context, phone string) (*entities.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (repository *userRepository) Create(ctx context.Context, user *entities.User) error {
	return repository.db.WithContext(ctx).Create(user).Error
}
func (repository *userRepository) Get(ctx context.Context) ([]*entities.User, error) {
	var users []*entities.User
	err := repository.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (repository *userRepository) GetID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := repository.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (repository *userRepository) Patch(ctx context.Context, user *entities.User) error {
	return repository.db.WithContext(ctx).Updates(user).Error
}

func (repository *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return repository.db.WithContext(ctx).Delete(&entities.User{}, "id = ?", id).Error
}

func (repository *userRepository) FindByPhone(ctx context.Context, phone string) (*entities.User, error) {
	var user entities.User
	err := repository.db.WithContext(ctx).First(&user, "phone = ?", phone).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
