package sso

import "errors"

var (
	ErrAppIDMustBePositive = errors.New("sso app id must be positive")
	ErrSsoAddressIsEmpty   = errors.New("sso address is empty")
)
