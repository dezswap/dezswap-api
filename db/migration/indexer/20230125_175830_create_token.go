//go:build mig
// +build mig

package main

import (
	"github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var (
	MainnetToken = indexer.Token{
		ChainModel: indexer.ChainModel{
			ChainId: "dimension_37-1",
			Address: "axpla",
		},
		Protocol: "",
		Symbol:   "XPLA",
		Name:     "XPLA",
		Decimals: 18,
		Icon:     "https://assets.xpla.io/icon/svg/XPLA.svg",
		Verified: true,
	}
	TestnetToken = indexer.Token{
		ChainModel: indexer.ChainModel{
			ChainId: "cube_47-5",
			Address: "axpla",
		},
		Protocol: "",
		Symbol:   "XPLA",
		Name:     "XPLA",
		Decimals: 18,
		Icon:     "https://assets.xpla.io/icon/svg/XPLA.svg",
		Verified: true,
	}
)

var M20230125_175830 = &gormigrate.Migration{
	ID: "20230125_175830",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&indexer.Token{}); err != nil {
			return err
		}
		if err := tx.Save([]indexer.Token{MainnetToken, TestnetToken}).Error; err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&indexer.Token{})
	},
}
