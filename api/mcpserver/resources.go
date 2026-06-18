package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const resourceMIMETypeMarkdown = "text/markdown"

type resourceSpec struct {
	URI         string
	Name        string
	Title       string
	Description string
	MIMEType    string
	Text        string
}

func defaultResources() []resourceSpec {
	return []resourceSpec{
		{
			URI:         "dezswap://service/guide",
			Name:        "dezswap-service-guide",
			Title:       "Dezswap Service Guide",
			Description: "Service boundaries and recommended MCP tool usage for Dezswap analytics.",
			MIMEType:    resourceMIMETypeMarkdown,
			Text: `# Dezswap Service Guide

This MCP server provides read-only Dezswap analytics and discovery. Use its tools to inspect market data, token metadata, pool state, route availability, charts, recent transactions, health, and version information.

## Boundaries

- The server cannot execute swaps.
- The server cannot build, sign, or submit transactions.
- The server cannot return amount-based quotes.
- The server cannot calculate slippage.
- The server cannot calculate price impact.
- The server cannot guarantee route execution.

## Intent-Level Tool Guidance

- For a market overview, use the market overview tool.
- For token metadata such as symbol, decimals, icon, protocol, or verified status, use the token metadata tool.
- For token market data such as price, TVL, volume, and fee, use the token market tool.
- For pools related to a token, use the pool listing tool with the token filter.
- For route availability, use the route discovery tool.
- For recent swap, add-liquidity, or remove-liquidity activity, use the recent transactions tool.

## Workflows

- Token basic analysis: get token metadata, then token market data, then related pools.
- Token trend analysis: only call token chart tools when the user asks for history, trend, or time-series data.
- Pool basic analysis: get pool metadata, then pool detail.
- Pool trend analysis: only call pool chart tools when the user asks for history, trend, or time-series data.
- Route availability: use route discovery and answer only whether a path candidate exists. Do not describe the result as a quote, expected output, execution guarantee, slippage estimate, or price impact estimate.

## Route Caveat

The hopCount argument is a maximum path length filter. It is not a swap amount, quote parameter, slippage setting, or price impact parameter. Do not treat an omitted hopCount as unlimited; set hopCount explicitly when looking for multi-hop routes.
`,
		},
		{
			URI:         "dezswap://dex/domain-model",
			Name:        "dezswap-domain-model",
			Title:       "Dezswap Domain Model",
			Description: "Dezswap-specific concepts, address conventions, and response interpretation rules.",
			MIMEType:    resourceMIMETypeMarkdown,
			Text: `# Dezswap Domain Model

## Core Concepts

- Token: an asset traded or tracked by Dezswap. A token may be a native denom, an IBC denom, a CW20 token, or an ERC20-prefixed asset.
- Pair: a swap contract connecting two assets, plus a liquidity provider token.
- Pool: the current liquidity state for a pair, including asset amounts and total share.
- LP token: the liquidity provider share token for a pair. It is not one of the two swapped assets.
- Route: a precomputed token path candidate. It is not a quote, transaction, execution guarantee, expected output, slippage estimate, or price impact estimate.
- Chart: a raw time series with {t, v} items.
- Transaction action: public transaction action values are swap, add, and remove. The add and remove values correspond to internal provide and withdraw events.

## Address Conventions

- Native assets can look like axpla.
- IBC assets are stored as ibc/<hash>.
- For path parameters, encode ibc/<hash> as ibc-<hash> so the slash does not split the URL path.
- XPLA token identifiers can include prefixed CW20 or ERC20 denom forms.

## Numeric Values

Prices, TVL, volume, fees, liquidity, total share, token amounts, and similar quantities are returned as decimal strings. Do not parse them as binary floating-point values when precision matters.

## Chart Interpretation

- t is a UTC timestamp.
- v is the raw value for the requested chart type.
- Chart responses do not include symbol, unit, computed summary, quote, confidence, slippage, or price impact.
- Use charts only for history or trend questions. They are not needed for basic token or pool analysis.

## Verified Tokens

The verified field means the token matched configured or known asset metadata. It is not investment safety, liquidity quality, route availability, execution success, or price reliability advice.
`,
		},
	}
}

func addResources(server *mcp.Server) {
	for _, spec := range defaultResources() {
		spec := spec
		server.AddResource(&mcp.Resource{
			URI:         spec.URI,
			Name:        spec.Name,
			Title:       spec.Title,
			Description: spec.Description,
			MIMEType:    spec.MIMEType,
			Size:        int64(len(spec.Text)),
		}, readStaticResource(spec))
	}
}

func readStaticResource(spec resourceSpec) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := spec.URI
		if req != nil && req.Params != nil {
			uri = req.Params.URI
		}
		if uri != spec.URI {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      spec.URI,
					MIMEType: spec.MIMEType,
					Text:     spec.Text,
				},
			},
		}, nil
	}
}
