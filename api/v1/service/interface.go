package service

import (
	"context"
	"errors"
)

// ErrInvalidKey is what a Getter returns for a key that is not one it can serve.
// A handler answers it with 400 instead of reporting it: keys are read from the
// URL, so a caller walking made-up ones would otherwise raise one alert per
// request.
var ErrInvalidKey = errors.New("invalid key")

type Getter[T any] interface {
	Get(key string) (*T, error)
	GetAll() ([]T, error)
}

type StatusService interface {
	CheckDB(ctx context.Context) error
	CheckCache(ctx context.Context) error
}
