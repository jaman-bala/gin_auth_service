package services

import (
	"context"
	"fmt"
	"gold_portal/internal/errors"
	"mime/multipart"
	"time"

	"github.com/google/uuid"

	"gold_portal/internal/domain/dto"
	"gold_portal/internal/domain/repositories"
)

type UsersService interface {
	GetAll(ctx context.Context) ([]*dto.UserResponseDTO, error)
	UserID(ctx context.Context, id uuid.UUID) (*dto.UserResponseDTO, error)
	GetByPhone(ctx context.Context, phone string) (*dto.UserResponseDTO, error)
	Patch(ctx context.Context, id uuid.UUID, request dto.UserUpdateDTO, photoFile *multipart.FileHeader) (*dto.UserResponseDTO, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	usersRepository repositories.UserRepository
	fileService     FileService
}

func NewUserService(usersRepository repositories.UserRepository, fileService FileService) UsersService {
	return &userService{
		usersRepository: usersRepository,
		fileService:     fileService,
	}
}

func (s *userService) GetAll(ctx context.Context) ([]*dto.UserResponseDTO, error) {
	users, err := s.usersRepository.Get(ctx)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	var response []*dto.UserResponseDTO
	for _, user := range users {
		var userResponse dto.UserResponseDTO
		userResponse.FromModel(user)
		response = append(response, &userResponse)
	}
	return response, nil
}

func (s *userService) UserID(ctx context.Context, id uuid.UUID) (*dto.UserResponseDTO, error) {
	user, err := s.usersRepository.GetID(ctx, id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	var response dto.UserResponseDTO
	response.FromModel(user)
	return &response, nil
}

func (s *userService) GetByPhone(ctx context.Context, phone string) (*dto.UserResponseDTO, error) {
	user, err := s.usersRepository.FindByPhone(ctx, phone)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	var userResponse dto.UserResponseDTO
	userResponse.FromModel(user)
	return &userResponse, nil
}

func (s *userService) Patch(ctx context.Context, id uuid.UUID, request dto.UserUpdateDTO, photoFile *multipart.FileHeader) (*dto.UserResponseDTO, error) {
	if id == uuid.Nil {
		return nil, errors.ErrInvalidUUID
	}
	user, err := s.usersRepository.GetID(ctx, id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	request.ApplyToModel(user)

	// Если есть фото, сохраняем его
	if photoFile != nil {
		filePath, err := s.saveUserPhoto(ctx, photoFile)
		if err != nil {
			return nil, err
		}
		user.Photo = filePath
	}

	user.UpdatedAt = time.Now()
	if err := s.usersRepository.Patch(ctx, user); err != nil {
		return nil, errors.ErrUpdateConflict
	}
	var response dto.UserResponseDTO
	response.FromModel(user)
	return &response, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.ErrInvalidUUID
	}
	return s.usersRepository.Delete(ctx, id)
}

// saveUserPhoto сохраняет фото пользователя в MinIO
func (s *userService) saveUserPhoto(ctx context.Context, photoFile *multipart.FileHeader) (string, error) {
	if photoFile == nil {
		return "", nil
	}

	filePath, err := s.fileService.UploadFile(ctx, photoFile, "user")
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении фото: %w", err)
	}

	return filePath, nil
}
