package parser

// source models from parser and aggregator
import (
	"github.com/dezswap/cosmwasm-etl/parser"
	"github.com/dezswap/cosmwasm-etl/pkg/db/schemas"
)

type Meta map[string]interface{}

type TxType parser.TxType

type Pair struct {
	ID string `json:"id" gorm:"primarykey"`
	schemas.Pair
}

type PoolInfo = schemas.PoolInfo

type ParsedTx struct {
	ID uint64 `json:"id" gorm:"primarykey"`
	schemas.ParsedTx
}

type SyncedHeight = schemas.SyncedHeight
