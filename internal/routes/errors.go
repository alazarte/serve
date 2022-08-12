package routes

import (
	"errors"
)

var (
	ErrFileNotFound        = errors.New("File not found")
	ErrInternalServerError = errors.New("Internal server error")
)
