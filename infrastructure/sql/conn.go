package sql

import (
	"database/sql"
	"embed"
	"errors"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"

	// side effect to load sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations
var migrations embed.FS

func SetupSQLConnection(databaseURL string) *sqlx.DB {
	driverName := "sqlite3"
	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	sourceInstance, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance(
		"httpfs", sourceInstance, "sqlite", driver)

	if err != nil {
		log.Fatalln("err preparing migrations", err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalln("err executing migrations", err)
	}
	log.Println("sql: all migrations run successfully")

	dbSQLX := sqlx.NewDb(db, driverName)
	if err := dbSQLX.Ping(); err != nil {
		log.Fatalln("err pinging conn", err)
	}
	return dbSQLX
}
