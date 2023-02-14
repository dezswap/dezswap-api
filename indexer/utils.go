package indexer

type comparable interface {
	Equal(comparable) bool
}

func isEqual(a, b comparable) bool {
	return a.Equal(b)
}
