package service

import "context"

type Getter[T any] interface {
	Get(key string) (*T, error)
	GetAll() ([]T, error)
}

type StatusService interface {
	CheckDB(ctx context.Context) error
	CheckCache(ctx context.Context) error
}
