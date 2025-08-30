package services

import (
	"context"
	"fmt"
	"gold_portal/config"
	"gold_portal/internal/domain/entities"
	"gold_portal/internal/domain/repositories"
	"gold_portal/internal/errors"
	"mime/multipart"
	"time"

	"gold_portal/internal/domain/dto"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type AuthService interface {
	UserRegister(ctx context.Context, request dto.UserRequestDTO, photoFile *multipart.FileHeader) (*dto.UserResponseDTO, error)
	Register(ctx context.Context, request dto.UserDashboardDTO, photoFile *multipart.FileHeader) (*dto.UserResponseDTO, error)
	Login(ctx context.Context, request dto.LoginRequestDTO) (*dto.LoginResponseDTO, error)
	Logout(ctx context.Context, tokenString string) error
	UserMe(ctx context.Context, id uuid.UUID) (*dto.UserResponseDTO, error)
	VerifyToken(tokenString string) (*jwt.Token, error)
	GetUserFromToken(ctx context.Context, token *jwt.Token) (*dto.UserResponseDTO, error)
	GenerateRefreshToken(user *entities.User) (string, time.Time, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponseDTO, error)
	GetAccessTokenExpiry() time.Duration
	GetRefreshTokenExpiry() time.Duration
}

type authService struct {
	userRepository repositories.UserRepository
	tokenService   TokenService
	fileService    FileService
	config         *config.Config
}

func NewAuthService(userRepository repositories.UserRepository, tokenService TokenService, fileService FileService, config *config.Config) AuthService {
	return &authService{
		userRepository: userRepository,
		tokenService:   tokenService,
		fileService:    fileService,
		config:         config,
	}
}

func (s *authService) GetAccessTokenExpiry() time.Duration {
	return s.config.JWT.Expiry
}

func (s *authService) GetRefreshTokenExpiry() time.Duration {
	return s.config.JWT.RefreshExpiry
}

func (s *authService) generateJWTToken(user *entities.User) (accessToken string, refreshToken string, expiresAt time.Time, err error) {
	accessExpiry := time.Now().Add(s.config.JWT.Expiry)
	refreshExpiry := time.Now().Add(s.config.JWT.RefreshExpiry)

	accessClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     accessExpiry.Unix(),
		"type":    "access",
		"jti":     uuid.New().String(),
		"iat":     time.Now().Unix(),
	}

	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     refreshExpiry.Unix(),
		"type":    "refresh",
		"jti":     uuid.New().String(),
		"iat":     time.Now().Unix(),
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessToken, err = at.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("ошибка подписи access токена: %w", err)
	}

	refreshToken, err = rt.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("ошибка подписи refresh токена: %w", err)
	}

	return accessToken, refreshToken, accessExpiry, nil
}

func (s *authService) VerifyToken(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, errors.ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["type"] != "access" {
			return nil, errors.ErrInvalidToken
		}
	}

	return token, nil
}

func (s *authService) GenerateRefreshToken(user *entities.User) (string, time.Time, error) {
	if s.config.JWT.Secret == "" {
		return "", time.Time{}, errors.ErrTokenConfig
	}
	expiresAt := time.Now().Add(s.config.JWT.RefreshExpiry)

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     expiresAt.Unix(),
		"type":    "refresh",
		"jti":     uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *authService) GetUserFromToken(ctx context.Context, token *jwt.Token) (*dto.UserResponseDTO, error) {
	if token == nil || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrEXP
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.ErrJTI
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге ID пользователя: %w", err)
	}

	user, err := s.userRepository.GetID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	var userResp dto.UserResponseDTO
	userResp.FromModel(user)
	return &userResp, nil
}

func (s *authService) UserRegister(ctx context.Context, request dto.UserRequestDTO, photoFile *multipart.FileHeader) (*dto.UserResponseDTO, error) {
	existingUser, err := s.userRepository.FindByPhone(ctx, request.Phone)
	if err == nil && existingUser != nil {
		return nil, errors.ErrUserPhoneExists
	}

	user := request.ToModelUser(request.Password)
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if photoFile != nil {
		filePath, err := s.saveUserPhoto(ctx, photoFile)
		if err != nil {
			return nil, err
		}
		user.Photo = filePath
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	var userResponse dto.UserResponseDTO
	userResponse.FromModelUser(user)

	return &userResponse, nil
}

func (s *authService) Register(ctx context.Context, req dto.UserDashboardDTO, photoFile *multipart.FileHeader) (*dto.UserResponseDTO, error) {
	existingUser, err := s.userRepository.FindByPhone(ctx, req.Phone)
	if err == nil && existingUser != nil {
		return nil, errors.ErrUserPhoneExists
	}

	user := req.ToModel(req.Password)
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if photoFile != nil {
		filePath, err := s.saveUserPhoto(ctx, photoFile)
		if err != nil {
			return nil, err
		}
		user.Photo = filePath
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	var userResponse dto.UserResponseDTO
	userResponse.FromModel(user)

	return &userResponse, nil
}

func (s *authService) Login(ctx context.Context, request dto.LoginRequestDTO) (*dto.LoginResponseDTO, error) {
	user, err := s.userRepository.FindByPhone(ctx, request.Phone)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, errors.ErrAccountBlocked
	}
	if err := user.CheckPassword(request.Password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}
	accessToken, refreshToken, _, err := s.generateJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("token generation error: %w", err)
	}
	var userResponse dto.UserResponseDTO
	userResponse.FromModel(user)

	tokenResponse := &dto.LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Success authorization",
	}
	return tokenResponse, nil
}

func (s *authService) Logout(ctx context.Context, tokenString string) error {
	if tokenString == "" {
		return errors.ErrInvalidToken
	}

	// Парсим токен для получения времени истечения
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return errors.ErrInvalidToken
	}

	// Получаем время истечения токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.ErrInvalidToken
	}

	expiryTime := time.Unix(int64(exp), 0)

	// Добавляем токен в черный список
	err = s.tokenService.BlacklistToken(ctx, tokenString, expiryTime)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (s *authService) UserMe(ctx context.Context, id uuid.UUID) (*dto.UserResponseDTO, error) {
	user, err := s.userRepository.GetID(ctx, id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	var userResp dto.UserResponseDTO
	userResp.FromModel(user)
	return &userResp, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponseDTO, error) {
	if refreshToken == "" {
		return nil, errors.ErrInvalidToken
	}

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrInvalidToken
	}

	if claims["type"] != "refresh" {
		return nil, errors.ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.ErrInvalidToken
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге ID пользователя: %w", err)
	}

	user, err := s.userRepository.GetID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	accessToken, newRefreshToken, expiresAt, err := s.generateJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("token generation error: %w", err)
	}
	_ = newRefreshToken

	var userResp dto.UserResponseDTO
	userResp.FromModel(user)

	return &dto.TokenResponseDTO{
		AccessToken: accessToken,
		User:        userResp,
		ExpiresAt:   expiresAt,
		Message:     "Token refreshed successfully",
	}, nil
}

// saveUserPhoto сохраняет фото пользователя в MinIO
func (s *authService) saveUserPhoto(ctx context.Context, photoFile *multipart.FileHeader) (string, error) {
	if photoFile == nil {
		return "", nil
	}

	filePath, err := s.fileService.UploadFile(ctx, photoFile, "user")
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении фото: %w", err)
	}

	return filePath, nil
}
