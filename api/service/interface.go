package service

type Gettable interface {
	Pool | Token | Pair
}

type Getter[T Gettable] interface {
	Get(key string) (T, error)
	GetAll() ([]T, error)
}
