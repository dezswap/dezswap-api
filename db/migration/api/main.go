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

const (
	modeUp   = "up"
	modeDown = "down"
)

var migrations = []*gormigrate.Migration{M20231121_201814}

func main() {
	mode := modeUp
	if len(os.Args) > 2 {
		log.Fatalf("usage: %s [up|down]", os.Args[0])
	}
	if len(os.Args) == 2 {
		mode = os.Args[1]
	}
	if mode != modeUp && mode != modeDown {
		log.Fatalf("unknown command: %q (expected %q or %q)", mode, modeUp, modeDown)
	}
	c := configs.New().Api.DB

	url := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s", c.Username, c.Password, c.Host, c.Port, c.Database)
	if c.SSLMode != "" {
		url = fmt.Sprintf("%s sslmode=%s", url, c.SSLMode)
	}
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

	if mode == modeDown {
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
