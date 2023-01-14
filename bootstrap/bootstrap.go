package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/togglhire/backend-homework/config"
	"github.com/togglhire/backend-homework/infrastructure/server"
	"github.com/togglhire/backend-homework/infrastructure/sql"

	"github.com/togglhire/backend-homework/usecase"
)

type Closer interface {
	Close() error
}

func CreateServerAndDependencies(cfg config.Config) (context.Context, *server.Server, []Closer, error) {

	// INFRA
	db := sql.SetupSQLConnection(cfg.DatabaseUrl)
	repo := sql.NewRepo(db)

	// USECASE
	questions := usecase.NewQuestions(repo)

	// SERVER
	ctx, srv := server.NewServer(context.Background(), cfg.Port, questions)
	return ctx, srv, []Closer{srv, db}, nil
}

func Run() error {

	cfg := config.Parse()
	ctx, srv, closers, err := CreateServerAndDependencies(cfg)
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
