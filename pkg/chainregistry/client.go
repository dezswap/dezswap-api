package chainregistry

import (
	"fmt"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/types"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

const mainnetAssetlistEndpointFormat = "https://raw.githubusercontent.com/cosmos/chain-registry/refs/heads/master/%s/assetlist.json"
const testnetAssetlistEndpointFormat = "https://raw.githubusercontent.com/cosmos/chain-registry/refs/heads/master/testnets/%s/assetlist.json"

// key: chain ID prefix
var assetListEndpointMap = map[string]string{
	"fetchhub": fmt.Sprintf(mainnetAssetlistEndpointFormat, "fetchhub"),
	"dorado":   fmt.Sprintf(testnetAssetlistEndpointFormat, "fetchhubtestnet"),
	"columbus": fmt.Sprintf(mainnetAssetlistEndpointFormat, "terra"),
	"phoenix":  fmt.Sprintf(mainnetAssetlistEndpointFormat, "terra2"),
	"pisco":    fmt.Sprintf(testnetAssetlistEndpointFormat, "terra2testnet"),
}

type client struct {
	http.Client
	AssetListEndpoint string
}

var _ pkg.Client = &client{}

func NewClient(chainId string) (pkg.Client, error) {
	for k, v := range assetListEndpointMap {
		if strings.HasPrefix(chainId, k) {
			return &client{http.Client{}, v}, nil
		}
	}

	return nil, errors.New("unsupported network")
}

// VerifiedCw20s implements Client
func (c *client) VerifiedCw20s() (*types.TokensRes, error) {
	res, err := pkg.GetAndUnmarshal[types.AssetsRes](&c.Client, c.AssetListEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedCw20s")
	}

	converted := types.TokensRes{
		Mainnet: types.TokenResMap{},
		Testnet: types.TokenResMap{},
	}

	if res.Assets != nil {
		for _, a := range res.Assets {
			if a.TypeAsset == nil || *a.TypeAsset != string(types.AssetTypeCw20) {
				continue
			}

			converted.Mainnet[*a.Address] = types.TokenRes{
				Symbol:   a.Symbol,
				Name:     a.Name,
				Token:    a.Address,
				Icon:     getIcon(a),
				Decimals: getDecimals(a),
			}
		}
	}

	return &converted, nil
}

// VerifiedIbcs implements Client
func (c *client) VerifiedIbcs() (*types.IbcsRes, error) {
	res, err := pkg.GetAndUnmarshal[types.AssetsRes](&c.Client, c.AssetListEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "VerifiedIbcs")
	}

	converted := types.IbcsRes{
		Mainnet: types.IbcResMap{},
		Testnet: types.IbcResMap{},
	}

	if res.Assets != nil {
		for _, a := range res.Assets {
			if a.TypeAsset == nil || *a.TypeAsset != string(types.AssetTypeIcs20) {
				continue
			}

			baseDenom, path := getBaseDenomAndIbcPath(a)
			converted.Mainnet[*a.Base] = types.IbcRes{
				Denom:     a.Base,
				Path:      path,
				BaseDenom: baseDenom,
				Symbol:    a.Symbol,
				Name:      a.Name,
				Icon:      getIcon(a),
				Decimals:  getDecimals(a),
			}
		}
	}

	return &converted, nil
}

func (c *client) VerifiedErc20s() (*types.TokensRes, error) {
	return &types.TokensRes{}, nil
}

func getDecimals(asset types.AssetRes) *uint8 {
	decimals := uint8(0)
	for _, du := range asset.DenomUnits {
		if decimals < du.Exponent {
			decimals = du.Exponent
		}
	}

	return &decimals
}

func getIcon(asset types.AssetRes) *string {
	var icon *string
	for _, uri := range asset.LogoUris {
		icon = &uri // pick one if any
	}

	return icon
}

func getBaseDenomAndIbcPath(asset types.AssetRes) (*string, *string) {
	for _, t := range asset.Traces {
		if t.Type != nil && (*t.Type == string(types.TraceTypeIbc) || *t.Type == string(types.TraceTypeIbcCw20)) {
			return t.CounterParty.BaseDenom, t.Chain.Path
		}
	}

	return nil, nil
}
