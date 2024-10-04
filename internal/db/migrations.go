package db

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"path/filepath"
)

func RunMigrations(db *sql.DB) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Could not create a migration driver: %v", err)
	}
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("Could not get migrations path: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"mysql", driver)
	if err != nil {
		log.Fatalf("Could not create a migration instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Could not apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}
