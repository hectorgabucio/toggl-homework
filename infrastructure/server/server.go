package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const SRV_SHUTDOWN_TIMEOUT = 10

type Server struct {
	port int
	srv  *http.Server
}

func NewServer(ctx context.Context, port int) (context.Context, *Server) {
	srv := Server{port: port, srv: &http.Server{Addr: fmt.Sprintf(":%d", port)}}
	return serverContext(ctx), &srv
}

func (server *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*SRV_SHUTDOWN_TIMEOUT)
	defer cancel()
	return server.srv.Shutdown(ctx)
}

func (server *Server) Run(ctx context.Context) error {
	log.Println("HTTP server starting on port", server.port)
	go func() {
		if err := server.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("err on http server: %s", err)
		}
	}()

	<-ctx.Done()
	if err := ctx.Err(); err != nil && err != context.Canceled {
		return fmt.Errorf("reason why context canceled, %w", err)
	}
	return nil
}

func serverContext(ctx context.Context) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-c
		cancel()
	}()

	return ctx
}
