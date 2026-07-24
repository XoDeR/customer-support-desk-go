package main

import (
	"context"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/config"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/database"
	"github.com/XoDeR/customer-support-desk-go/pkg/migration"
	"log"
)

func main() {
	c, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	db, e := database.NewPostgresConnection(c)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if e = migration.NewManager(db, "migrations/app").Migrate(context.Background()); e != nil {
		log.Fatal(e)
	}
}
