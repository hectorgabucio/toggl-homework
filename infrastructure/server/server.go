package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/togglhire/backend-homework/usecase"
)

const SRV_SHUTDOWN_TIMEOUT = 10

type Server struct {
	port      int
	srv       *http.Server
	questions usecase.Questions
}

func NewServer(ctx context.Context, port int, questions usecase.Questions) (context.Context, *Server) {
	srv := Server{port: port, srv: &http.Server{Addr: fmt.Sprintf(":%d", port)}, questions: questions}
	return serverContext(ctx), &srv
}

func (server *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*SRV_SHUTDOWN_TIMEOUT)
	defer cancel()
	return server.srv.Shutdown(ctx)
}

func (s *Server) Run(ctx context.Context) error {
	log.Println("HTTP server starting on port", s.port)

	http.HandleFunc("/status", s.handleStatus)
	http.HandleFunc("/questions", s.handleListQuestions)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("err on http server: %s", err)
		}
	}()

	<-ctx.Done()
	if err := ctx.Err(); err != nil && err != context.Canceled {
		return fmt.Errorf("reason why context canceled, %w", err)
	}
	return nil
}

func (s Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s Server) handleListQuestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
		return
	}

	questions := s.questions.GetAll()

	w.Header().Add("Content-Type", "application/json")

	json.NewEncoder(w).Encode(questions)
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
