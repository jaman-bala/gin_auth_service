package errors

import "errors"

var (
	ErrInvalidUserRole = errors.New("invalid user role")
	ErrUserNotFound    = errors.New("user not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUserExists      = errors.New("user already exists")
	ErrUserPhoneExists = errors.New("user phone already exists")
	ErrInvalidUUID     = errors.New("invalid uuid")
)

var (
	ErrUpdateConflict = errors.New("update conflict")
)

var (
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrAccountBlocked     = errors.New("account blocked")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrEXP          = errors.New("expired token")
	ErrJTI          = errors.New("invalid jti")
	ErrTokenConfig  = errors.New("token config error")
)
