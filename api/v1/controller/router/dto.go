package router

type RoutesRes []RouteRes

type RouteRes struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	HopCount int      `json:"hopCount"`
	Route    []string `json:"route"`
}
