package service

type Gettable interface {
}

type GetterService[T Gettable] interface {
	Get() (T, error)
	GetAll() ([]T, error)
}
