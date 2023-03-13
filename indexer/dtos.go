package indexer

type Updatable interface {
	Token | PoolInfo
}

type Action string

const (
	Swap     Action = "swap"
	Provide  Action = "provide"
	Withdraw Action = "withdraw"
)

type Pair struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Asset0  string `json:"asset0"`
	Asset1  string `json:"asset1"`
	Lp      string `json:"lp"`
}

type Token struct {
	ID       uint   `json:"id"`
	Address  string `json:"address"`
	ChainId  string `json:"chainId"`
	Protocol string `json:"protocol"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
	Icon     string `json:"icon"`
	Verified bool   `json:"verified"`
}

// Equal implements comparable
func (lhs *Token) Equal(rhs comparable) bool {
	t, ok := rhs.(*Token)
	if !ok {
		return false
	}
	return lhs.Address == t.Address &&
		lhs.Protocol == t.Protocol &&
		lhs.Symbol == t.Symbol &&
		lhs.Name == t.Name &&
		lhs.Decimals == t.Decimals &&
		lhs.Icon == t.Icon &&
		lhs.Verified == t.Verified
}

type PoolInfo struct {
	Height       uint64 `json:"height"`
	ChainId      string `json:"chainId"`
	Address      string `json:"address"`
	Asset0       string `json:"asset0"`
	Asset0Amount string `json:"asset0Amount"`
	Asset1       string `json:"asset1"`
	Asset1Amount string `json:"asset1Amount"`
	LpAmount     string `json:"lpAmount"`
}

// Equal implements comparable
func (lhs *PoolInfo) Equal(rhs comparable) bool {
	p, ok := rhs.(*PoolInfo)
	if !ok {
		return false
	}
	//!NOTE: doesn't have to compare height if amount of the token is the same
	return lhs.Address == p.Address &&
		lhs.Asset0Amount == p.Asset0Amount &&
		lhs.Asset1Amount == p.Asset1Amount &&
		lhs.LpAmount == p.LpAmount &&
		lhs.ChainId == p.ChainId
}

type ParsedTx struct {
	ID                uint64
	ChainId           string  `json:"chainId"`
	Height            uint64  `json:"height"`
	Timestamp         float64 `json:"timestamp"` // timestamp of a block in second
	Hash              string  `json:"hash"`
	Sender            string  `json:"sender"`
	Type              Action  `json:"type"`
	Address           string  `json:"address"`
	Asset0            string  `json:"asset0"`
	Asset0Amount      string  `json:"asset0Amount"`
	Asset1            string  `json:"asset1"`
	Asset1Amount      string  `json:"asset1Amount"`
	Lp                string  `json:"lp"`
	LpAmount          string  `json:"lpAmount"`
	CommissionAmount  string  `json:"commissionAmount"`
	Commission0Amount string  `json:"commission0Amount"`
	Commission1Amount string  `json:"commission1Amount"`
}
