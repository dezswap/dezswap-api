//go:build mig
// +build mig

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var migrations = []*gormigrate.Migration{M20231121_201814}

func main() {
	rollback := os.Args[len(os.Args)-1]
	c := configs.New().Api.DB

	url := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", c.Username, c.Password, c.Host, c.Port, c.Database)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: url,
	}))
	if err != nil {
		panic(err)
	}
	m := gormigrate.New(db, &gormigrate.Options{
		TableName:                 "migrations_api",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            true,
		ValidateUnknownMigrations: false,
	}, migrations)

	if rollback == "down" {
		log.Printf("Migration Rollback is running...")
		if err := m.RollbackLast(); err != nil {
			panic(errors.Wrap(err, "Down"))
		}
		log.Printf("Rollback ran successfully")
		return
	}

	log.Printf("Migration is running...")
	if err = m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Printf("Migration did run successfully")
}
