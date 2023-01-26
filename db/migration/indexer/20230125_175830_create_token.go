//go:build mig
// +build mig

package main

import (
	"github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var M20230125_175830 = &gormigrate.Migration{
	ID: "20230125_175830",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&indexer.Token{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&indexer.Token{})
	},
}
