package dashboard

import dashboardService "github.com/dezswap/dezswap-api/api/service/dashboard"

type mapper struct{}

func (m *mapper) tokenToRes(token dashboardService.Token) TokenRes {
	return TokenRes{
		Address:         string(token.Addr),
		Price:           token.Price,
		PriceChange:     token.PriceChange,
		Volume24h:       token.Volume,
		Volume24hChange: token.VolumeChange,
		Volume7d:        token.Volume7d,
		Volume7dChange:  token.Volume7dChange,
		Tvl:             token.Tvl,
		TvlChange:       token.TvlChange,
	}
}

func (m *mapper) tokensToRes(tokens dashboardService.Tokens) TokensRes {
	res := make(TokensRes, len(tokens))
	for i, t := range tokens {
		res[i] = m.tokenToRes(t)
	}

	return res
}
