package db

type Meta map[string]interface{}

type TxType string

const (
	CreatePair TxType = "create_pair"
	Swap       TxType = "swap"
	Provide    TxType = "provide"
	Withdraw   TxType = "withdraw"
	Transfer   TxType = "transfer"
)
