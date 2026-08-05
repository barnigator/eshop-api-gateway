package domain

import "errors"

var (
	ErrEmailRequired      = errors.New("email is required")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrPasswordRequired   = errors.New("password is required")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
