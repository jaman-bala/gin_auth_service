package services

import (
	"context"
	"fmt"
	"gold_portal/internal/pkg/jwt"
	"time"

	jwtv4 "github.com/golang-jwt/jwt/v4"
)

type TokenService interface {
	// Добавляет токен в черный список
	BlacklistToken(ctx context.Context, tokenString string, expiry time.Time) error
	// Проверяет, находится ли токен в черном списке
	IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error)
	// Удаляет токен из черного списка (например, при обновлении)
	RemoveFromBlacklist(ctx context.Context, tokenString string) error
	// Получает информацию о токене
	GetTokenInfo(ctx context.Context, tokenString string) (*jwt.TokenInfo, error)
}

type tokenService struct {
	cache Cache
	jwt   jwt.JWTService
}

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

func NewTokenService(cache Cache, jwtService jwt.JWTService) TokenService {
	return &tokenService{
		cache: cache,
		jwt:   jwtService,
	}
}

func (s *tokenService) BlacklistToken(ctx context.Context, tokenString string, expiry time.Time) error {
	if tokenString == "" {
		return fmt.Errorf("token string is required")
	}

	// Генерируем уникальный ключ для токена
	tokenKey := fmt.Sprintf("blacklist:%s", tokenString)

	// Вычисляем время жизни в черном списке
	// Токен должен оставаться в черном списке до истечения срока действия
	blacklistExpiry := time.Until(expiry)
	if blacklistExpiry <= 0 {
		blacklistExpiry = time.Hour // Минимум 1 час
	}

	// Сохраняем в кэш с временем жизни
	err := s.cache.Set(ctx, tokenKey, "blacklisted", blacklistExpiry)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

func (s *tokenService) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	if tokenString == "" {
		return false, fmt.Errorf("token string is required")
	}

	tokenKey := fmt.Sprintf("blacklist:%s", tokenString)

	exists, err := s.cache.Exists(ctx, tokenKey)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist status: %w", err)
	}

	return exists, nil
}

func (s *tokenService) RemoveFromBlacklist(ctx context.Context, tokenString string) error {
	if tokenString == "" {
		return fmt.Errorf("token string is required")
	}

	tokenKey := fmt.Sprintf("blacklist:%s", tokenString)

	err := s.cache.Delete(ctx, tokenKey)
	if err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	return nil
}

func (s *tokenService) GetTokenInfo(ctx context.Context, tokenString string) (*jwt.TokenInfo, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string is required")
	}

	// Парсим токен для получения информации
	token, err := s.jwt.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Проверяем, не в черном ли списке
	isBlacklisted, err := s.IsTokenBlacklisted(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist status: %w", err)
	}

	if isBlacklisted {
		return nil, fmt.Errorf("token is blacklisted")
	}

	// Извлекаем информацию из токена
	claims, ok := token.Claims.(jwtv4.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Приводим к правильному типу
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role in token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type in token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid exp in token")
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid iat in token")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid jti in token")
	}

	tokenInfo := &jwt.TokenInfo{
		UserID:    userID,
		Role:      role,
		Type:      tokenType,
		ExpiresAt: time.Unix(int64(exp), 0),
		IssuedAt:  time.Unix(int64(iat), 0),
		JTI:       jti,
	}

	return tokenInfo, nil
}
