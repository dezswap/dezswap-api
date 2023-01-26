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
	Address  string `json:"address"`
	Protocol string `json:"protocol"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
	Icon     string `json:"icon"`
	Verified bool   `json:"verified"`
}

type PoolInfo struct {
	Height       uint64 `json:"height"`
	Address      string `json:"address"`
	Asset0Amount string `json:"asset0Amount"`
	Asset1Amount string `json:"asset1Amount"`
	LpAmount     string `json:"lpAmount"`
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
