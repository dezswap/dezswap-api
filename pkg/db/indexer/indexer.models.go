package indexer

import (
	"gorm.io/gorm"
)

type ChainModel struct {
	ChainId string `json:"chainId" gorm:"not null;index:,unique,composite:chain_id_address_key"`
	Address string `json:"address" gorm:"not null;index:,unique,composite:chain_id_address_key"`
}

type Token struct {
	*gorm.Model
	ChainModel
	Protocol string `json:"protocol"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
	Icon     string `json:"icon"`
	Verified bool   `json:"verified" gorm:"not null;default:false"`
}

type LatestPool struct {
	*gorm.Model
	ChainModel
	Height       uint64 `json:"height"`
	Asset0       string `json:"asset0"`
	Asset0Amount string `json:"asset0Amount"`
	Asset1       string `json:"asset1"`
	Asset1Amount string `json:"asset1Amount"`
	LpAmount     string `json:"lpAmount"`
}
