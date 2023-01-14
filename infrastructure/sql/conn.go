package sql

import (
	"embed"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// side effect to load sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations
var migrations embed.FS

func SetupSQLConnection(databaseURL string) *gorm.DB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel: logger.Silent, // Log level
		},
	)
	g, err := gorm.Open(sqlite.Open(databaseURL), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatal(err)
	}

	db, err := g.DB()
	if err != nil {
		log.Fatal(err)
	}

	_ = g.Callback().Create().Remove("gorm:update_time_stamp")
	_ = g.Callback().Update().Remove("gorm:update_time_stamp")
	_ = g.Callback().Delete().Remove("gorm:update_time_stamp")

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

	if err := db.Ping(); err != nil {
		log.Fatalln("err pinging conn", err)
	}
	return g
}
