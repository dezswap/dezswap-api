package mcpserver

import (
	"net/http"

	dashboardController "github.com/dezswap/dezswap-api/api/v1/controller/dashboard"
	dashboardService "github.com/dezswap/dezswap-api/api/v1/service/dashboard"
)

var (
	marketChartTypeEnum = []string{
		dashboardController.ChartTypeVolume,
		dashboardController.ChartTypeTvl,
		dashboardController.ChartTypeApr,
		dashboardController.ChartTypeFee,
	}
	tokenChartTypeEnum = []string{
		dashboardController.ChartTypeVolume,
		dashboardController.ChartTypeTvl,
		dashboardController.ChartTypePrice,
	}
	durationEnum = []string{
		string(dashboardService.All),
		string(dashboardService.Month),
		string(dashboardService.Quarter),
		string(dashboardService.Year),
	}
	txTypeEnum = []string{
		string(dashboardController.TX_TYPE_SWAP),
		string(dashboardController.TX_TYPE_ADD),
		string(dashboardController.TX_TYPE_REMOVE),
	}
)

// defaultTools returns the built-in Dezswap MCP tools.
func defaultTools() []toolSpec {
	return []toolSpec{
		{
			Name:        "get_market_overview",
			Title:       "Get Market Overview",
			Description: "Get Dezswap-wide market metrics such as TVL, volume, fees, APR, change rates, address count, and transaction count. This is analytics data only.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/recent",
		},
		{
			Name:        "list_tokens",
			Title:       "List Tokens",
			Description: "List token market data including price, 24h and 7d volume, TVL, and fee where available.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/tokens",
		},
		{
			Name:        "get_token_market",
			Title:       "Get Token Market",
			Description: "Get market stats for one token, including price, volume, TVL, and fee. Address may be a native or contract token address accepted by Dezswap.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/tokens/{address}",
			Params: []paramSpec{
				addressPath("Token address"),
			},
		},
		{
			Name:        "get_token_metadata",
			Title:       "Get Token Metadata",
			Description: "Get token metadata such as name, symbol, decimals, icon, protocol, verified status, and total supply.",
			Method:      http.MethodGet,
			Path:        "/v1/tokens/{address}",
			Params: []paramSpec{
				addressPath("Token address"),
			},
		},
		{
			Name:        "list_pools",
			Title:       "List Pools",
			Description: "List Dezswap pools with TVL, volume, fee, and APR. Optionally filter by token address.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/pools",
			Params: []paramSpec{
				stringQuery("token", "Optional token address to filter related pools"),
			},
		},
		{
			Name:        "get_pool_detail",
			Title:       "Get Pool Detail",
			Description: "Get pool detail, recent pool stats, and recent transactions for a Dezswap pool.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/pools/{address}",
			Params: []paramSpec{
				addressPath("Pool address"),
			},
		},
		{
			Name:        "get_pool_metadata",
			Title:       "Get Pool Metadata",
			Description: "Get pool assets, total share, and related pool metadata.",
			Method:      http.MethodGet,
			Path:        "/v1/pools/{address}",
			Params: []paramSpec{
				addressPath("Pool address"),
			},
		},
		{
			Name:        "get_pair_metadata",
			Title:       "Get Pair Metadata",
			Description: "Get pair contract, liquidity token, and asset info metadata.",
			Method:      http.MethodGet,
			Path:        "/v1/pairs/{address}",
			Params: []paramSpec{
				addressPath("Pair address"),
			},
		},
		{
			Name:        "get_market_chart",
			Title:       "Get Market Chart",
			Description: "Get Dezswap-wide chart data for volume, TVL, APR, or fee. This is historical analytics data, not a quote.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/chart/{type}",
			Params: []paramSpec{
				enumPath("type", "Chart type", marketChartTypeEnum),
				durationQuery(),
			},
		},
		{
			Name:        "get_token_chart",
			Title:       "Get Token Chart",
			Description: "Get token chart data for volume, TVL, or price. This is historical analytics data, not a quote.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/chart/tokens/{address}/{type}",
			Params: []paramSpec{
				addressPath("Token address"),
				enumPath("type", "Chart type", tokenChartTypeEnum),
				durationQuery(),
			},
		},
		{
			Name:        "get_pool_chart",
			Title:       "Get Pool Chart",
			Description: "Get pool chart data for volume, TVL, APR, or fee. This is historical analytics data, not a quote.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/chart/pools/{address}/{type}",
			Params: []paramSpec{
				addressPath("Pool address"),
				enumPath("type", "Chart type", marketChartTypeEnum),
				durationQuery(),
			},
		},
		{
			Name:        "find_routes",
			Title:       "Find Routes",
			Description: "Find available token route paths by token addresses. This does not calculate amount-based quotes, expected output, slippage, or price impact.",
			Method:      http.MethodGet,
			Path:        "/v1/routes",
			Params: []paramSpec{
				stringQuery("from", "Offer token address. Either from or to is required."),
				stringQuery("to", "Ask token address. Either from or to is required."),
				integerQuery("hopCount", "Optional maximum hop count", false),
			},
		},
		{
			Name:        "list_recent_transactions",
			Title:       "List Recent Transactions",
			Description: "List recent Dezswap transactions, optionally filtered by pool or token and transaction type. Pool and token filters are mutually exclusive.",
			Method:      http.MethodGet,
			Path:        "/v1/dashboard/txs",
			Params: []paramSpec{
				stringQuery("pool", "Optional pool address"),
				stringQuery("token", "Optional comma-separated token addresses"),
				enumQuery("type", "Optional transaction type", txTypeEnum, false),
			},
		},
		{
			Name:        "get_service_health",
			Title:       "Get Service Health",
			Description: "Get API service health and dependency status.",
			Method:      http.MethodGet,
			Path:        "/v1/health",
		},
		{
			Name:        "get_service_version",
			Title:       "Get Service Version",
			Description: "Get the current API application version.",
			Method:      http.MethodGet,
			Path:        "/v1/version",
		},
	}
}

// addressPath creates a required address path parameter.
func addressPath(description string) paramSpec {
	return paramSpec{Name: "address", In: "path", Type: "string", Description: description, Required: true}
}

// enumPath creates a required enum path parameter.
func enumPath(name, description string, values []string) paramSpec {
	return paramSpec{Name: name, In: "path", Type: "string", Description: description, Required: true, Enum: values}
}

// stringQuery creates an optional string query parameter.
func stringQuery(name, description string) paramSpec {
	return paramSpec{Name: name, In: "query", Type: "string", Description: description}
}

// integerQuery creates an integer query parameter.
func integerQuery(name, description string, required bool) paramSpec {
	return paramSpec{Name: name, In: "query", Type: "integer", Description: description, Required: required}
}

// enumQuery creates an enum query parameter.
func enumQuery(name, description string, values []string, required bool) paramSpec {
	return paramSpec{Name: name, In: "query", Type: "string", Description: description, Required: required, Enum: values}
}

// durationQuery creates the standard chart duration parameter.
func durationQuery() paramSpec {
	return enumQuery("duration", "Optional chart duration. Omit for all data.", durationEnum, false)
}
