package service

type Getter[T any] interface {
	Get(key string) (*T, error)
	GetAll() ([]T, error)
}

type StatusService interface {
}
