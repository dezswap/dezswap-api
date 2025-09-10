package xpla

import (
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/types"
	"net/http"

	"github.com/pkg/errors"
)

type client struct {
	http.Client
}

var _ pkg.Client = &client{}

func NewClient() pkg.Client {
	return &client{http.Client{}}
}

// VerifiedCw20s implements Client
func (c *client) VerifiedCw20s() (*types.TokensRes, error) {
	res, err := pkg.GetAndUnmarshal[types.TokensRes](&c.Client, "https://assets.xpla.io/cw20/tokens.json")
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedCw20s")
	}

	return res, nil
}

// VerifiedIbcs implements Client
func (c *client) VerifiedIbcs() (*types.IbcsRes, error) {
	res, err := pkg.GetAndUnmarshal[types.IbcsRes](&c.Client, "https://assets.xpla.io/ibc/tokens.json")
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedIbcs")
	}

	return res, nil
}

func (c *client) VerifiedErc20s() (*types.TokensRes, error) {
	res, err := pkg.GetAndUnmarshal[types.TokensRes](&c.Client, "https://assets.xpla.io/erc20/tokens.json")
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedErc20s")
	}

	return res, nil
}
