package apperr

import "errors"

var (
	ErrNotFound   = errors.New("Not found")
	ErrInvalidURL = errors.New("invalid URL")
)
