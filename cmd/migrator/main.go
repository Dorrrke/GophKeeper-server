package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath string

	flag.StringVar(&storagePath, "storage-path", "postgres://postgres:6406655@localhost:5432/postgres?sslmode=disable", "path to database")
	flag.StringVar(&migrationsPath, "migrations-path", "migrations", "path to migrations")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}
	migratePath := "file://" + migrationsPath

	m, err := migrate.New(
		migratePath,
		storagePath,
	)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
	fmt.Println("Migrations apply")
}
