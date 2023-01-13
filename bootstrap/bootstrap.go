package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/togglhire/backend-homework/config"
	"github.com/togglhire/backend-homework/infrastructure/server"
	"github.com/togglhire/backend-homework/usecase"

	"github.com/golang-migrate/migrate/v4"

	// side effect to add file support for db migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// side effect to load sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

type Closer interface {
	Close() error
}

func createServerAndDependencies() (context.Context, *server.Server, []Closer, error) {

	// CONFIG
	cfg := config.Parse()

	// USECASE
	questions := usecase.NewQuestions()

	// INFRA
	db := setupSQLConnection(cfg.DatabaseUrl)
	ctx, srv := server.NewServer(context.Background(), cfg.Port, questions)
	return ctx, srv, []Closer{srv, db}, nil
}

func Run() error {
	ctx, srv, closers, err := createServerAndDependencies()
	defer closeResources(closers)
	if err != nil {
		return err
	}
	if err = srv.Run(ctx); err != nil {
		return fmt.Errorf("err running server, %w", err)
	}
	return nil
}

func closeResources(closers []Closer) {
	log.Println("ending all resources gracefully...")
	for _, closer := range closers {
		err := closer.Close()
		if err != nil {
			log.Println("err ending resource:", err)
		}
	}
}

func setupSQLConnection(databaseURL string) *sqlx.DB {
	driverName := "sqlite3"
	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://infrastructure/sql/migrations",
		driverName, driver)
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
