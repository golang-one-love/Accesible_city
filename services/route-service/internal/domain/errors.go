package service

import "errors"

var (
	ErrGraphNotInitialized = errors.New("graph not initialized")
	ErrInvalidCoordinates  = errors.New("invalid coordinates")
)