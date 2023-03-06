package parser

// source models from parser and aggregator
import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type Meta map[string]interface{}

type TxType string

const (
	CreatePair TxType = "create_pair"
	Swap       TxType = "swap"
	Provide    TxType = "provide"
	Withdraw   TxType = "withdraw"
	Transfer   TxType = "transfer"
)

type Pair struct {
	ID       string `json:"id" gorm:"primarykey"`
	ChainId  string `json:"chainId"`
	Contract string `json:"contract"`
	Asset0   string `json:"asset0"`
	Asset1   string `json:"asset1"`
	Lp       string `json:"lp"`

	Meta Meta `json:"meta"`
}

type PoolInfo struct {
	ChainId      string `json:"chainId" gorm:"index:idx_chain_id,unique"`
	Height       uint64 `json:"height"`
	Contract     string `json:"contract"`
	Asset0Amount string `json:"asset0Amount"`
	Asset1Amount string `json:"asset1Amount"`
	LpAmount     string `json:"lpAmount"`

	Meta Meta `json:"meta"`
}

type ParsedTx struct {
	ID                uint64
	ChainId           string  `json:"chainId"`
	Height            uint64  `json:"height"`
	Timestamp         float64 `json:"timestamp"` // timestamp of a block in second
	Hash              string  `json:"hash"`
	Sender            string  `json:"sender"`
	Type              TxType  `json:"type"`
	Contract          string  `json:"contract"`
	Asset0            string  `json:"asset0"`
	Asset0Amount      string  `json:"asset0Amount"`
	Asset1            string  `json:"asset1"`
	Asset1Amount      string  `json:"asset1Amount"`
	Lp                string  `json:"lp"`
	LpAmount          string  `json:"lpAmount"`
	CommissionAmount  string  `json:"commissionAmount"`
	Commission0Amount string  `json:"commission0Amount"`
	Commission1Amount string  `json:"commission1Amount"`

	Meta Meta `json:"meta"`
}

type SyncedHeight struct {
	ID      uint64 `json:"id" gorm:"primarykey"`
	ChainId string `json:"chainId"`
	Height  uint64 `json:"height"`
}

func (ParsedTx) TableName() string {
	return "parsed_tx"
}
func (PoolInfo) TableName() string {
	return "pool_info"
}
func (Pair) TableName() string {
	return "pair"
}
func (SyncedHeight) TableName() string {
	return "synced_height"
}

func (Meta) GormDataType() string {
	return "json"
}

func (j *Meta) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := Meta{}
	err := json.Unmarshal(bytes, &result)
	*j = Meta(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j Meta) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}
