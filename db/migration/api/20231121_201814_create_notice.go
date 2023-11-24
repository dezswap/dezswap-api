//go:build mig
// +build mig

package main

import (
	"github.com/dezswap/dezswap-api/pkg/db/api"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var M20231121_201814 = &gormigrate.Migration{
	ID: "20231121_201814",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&api.Notice{}); err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&api.Notice{})
	},
}
