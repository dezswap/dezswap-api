package pkg

import (
	"encoding/json"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type Client interface {
	VerifiedCw20s() (*types.TokensRes, error)
	VerifiedIbcs() (*types.IbcsRes, error)
	VerifiedErc20s() (*types.TokensRes, error)
}

func GetAndUnmarshal[T types.Unmarshalable](c *http.Client, url string) (*T, error) {
	res, err := c.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "pkg.GetAndUnmarshal")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "pkg.GetAndUnmarshal")
	}

	t := new(T)
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, errors.Wrap(err, "pkg.GetAndUnmarshal")
	}

	return t, nil
}
