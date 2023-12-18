package api

import (
	rs "github.com/dezswap/dezswap-api/api/service/router"
	"github.com/dezswap/dezswap-api/pkg/db/aggregator"

	"gorm.io/gorm"
)

type routerDbRepoImpl struct {
	chainId string
	*mapper
	db *gorm.DB
}

type mapper struct{}

func NewRouterDbRepo(chainId string, db *gorm.DB) rs.Router {
	return &routerDbRepoImpl{chainId, &mapper{}, db}
}

// RoutesOfToken implements Router.
func (r *routerDbRepoImpl) RoutesOfToken(addr string, hopCount int, reverse bool) ([]rs.Route, error) {
	models := []aggregator.Route{}
	query := r.db.Model(&aggregator.Route{}).Where("chain_id = ?", r.chainId).Order("hop_count ASC")

	if !reverse {
		query = query.Select("asset1, hop_count, route").Where("asset0 = ? AND hop_count <= ?", addr, hopCount).Order("asset1 ASC")
	} else {
		query = query.Select("asset0, hop_count, route").Where("asset1 = ? AND hop_count <= ?", addr, hopCount).Order("asset0 ASC")
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.modelsToRoutes(models, reverse), nil
}

// Routes implements Router.
func (r *routerDbRepoImpl) Routes(from string, to string, hopCount int) ([]rs.Route, error) {
	models := []aggregator.Route{}
	query := r.db.Model(&aggregator.Route{}).
		Where("chain_id = ? AND asset0 = ? AND asset1 = ? AND hop_count <= ?", r.chainId, from, to, hopCount).
		Order("hop_count ASC").Order("asset1 ASC")

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	return r.modelsToRoutes(models, false), nil
}

func (m *mapper) modelsToRoutes(models []aggregator.Route, reverse bool) []rs.Route {
	routes := make([]rs.Route, len(models))
	for i, model := range models {
		routes[i] = *(m.modelToRoute(&model, reverse))
	}
	return routes
}

func (m *mapper) modelToRoute(model *aggregator.Route, reverse bool) *rs.Route {
	route := &rs.Route{
		To:       model.Asset1,
		HopCount: model.HopCount,
		Route:    model.Route,
	}

	if reverse {
		route.To = model.Asset0
	}
	return route
}
