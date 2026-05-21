package domain

import "errors"

var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource already exists")
	ErrBadRequest = errors.New("invalid request data")
)
