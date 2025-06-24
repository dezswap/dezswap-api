package router

type Route struct {
	To       string   `json:"to"`
	HopCount int      `json:"hopCount"`
	Route    []string `json:"route"`
}

type Router interface {
	RoutesOfToken(addr string, hopCount int, reverse bool) ([]Route, error)
	Routes(from, to string, hopCount int) ([]Route, error)
}

type RouterRepo interface {
	RoutesOfToken(addr string, hopCount int, reverse bool) ([]Route, error)
	Routes(from, to string, hopCount int) ([]Route, error)
}

type routerImpl struct {
	RouterRepo
}

func New(repo RouterRepo) Router {
	return &routerImpl{repo}
}
