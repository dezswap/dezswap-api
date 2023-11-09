package dashboard

import dashboardService "github.com/dezswap/dezswap-api/api/service/dashboard"

type mapper struct{}

func (m *mapper) tokensToRes(tokens dashboardService.Tokens) TokensRes {
	res := make(TokensRes, len(tokens))
	for i, t := range tokens {
		res[i] = TokenRes{
			Address:     string(t.Addr),
			Price:       t.Price,
			PriceChange: t.PriceChange,
			Volume:      t.Volume,
			Tvl:         t.Tvl,
		}
	}

	return res
}
