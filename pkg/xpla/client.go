package xpla

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Client interface {
	VerifiedCw20s() (*TokensRes, error)
	VerifiedIbcs() (*IbcsRes, error)
}

type client struct {
	http.Client
}

func NewClient() Client {
	return &client{http.Client{}}
}

// VerifiedCw20s implements Client
func (c *client) VerifiedCw20s() (*TokensRes, error) {
	res, err := get[TokensRes](&c.Client, "https://assets.xpla.io/cw20/tokens.json")
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedCw20s")
	}

	return res, nil
}

// VerifiedIbcs implements Client
func (c *client) VerifiedIbcs() (*IbcsRes, error) {
	res, err := get[IbcsRes](&c.Client, "https://assets.xpla.io/ibc/tokens.json")
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedIbcs")
	}

	return res, nil
}

func get[T unmarshalable](c *http.Client, url string) (*T, error) {
	res, err := c.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "xpla.get")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "xpla.get")
	}

	t := new(T)
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, errors.Wrap(err, "xpla.get")
	}

	return t, nil
}
