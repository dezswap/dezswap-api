package indexer

import (
	"gorm.io/gorm"
)

type ChainModel struct {
	ChainId string `json:"chainId" gorm:"index:,unique,composite:chain_id_address_key"`
	Address string `json:"address" gorm:"index:,unique,composite:chain_id_address_key"`
}

type Token struct {
	gorm.Model
	ChainModel
	Protocol string `json:"protocol"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
	Icon     string `json:"icon"`
	Verified bool   `json:"verified"`
}
