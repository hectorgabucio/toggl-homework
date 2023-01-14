package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/togglhire/backend-homework/domain"
	"github.com/togglhire/backend-homework/usecase"

	"github.com/go-playground/validator/v10"
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
	http.HandleFunc("/questions", s.handleQuestions)

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

func (s Server) handleQuestions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listQuestions(w, r)
	case http.MethodPost:
		s.addQuestion(w, r)
	case http.MethodPut:
		s.updateQuestion(w, r)
	default:
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
	}
}

func (s Server) listQuestions(w http.ResponseWriter, r *http.Request) {

	questions := s.questions.GetAll()

	w.Header().Add("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(questions)
	if err != nil {
		log.Println("err encoding json response list questions", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s Server) addQuestion(w http.ResponseWriter, r *http.Request) {

	var question domain.Question
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Incorrect media type", http.StatusUnsupportedMediaType)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&question)
	if err != nil {
		http.Error(w, "failed to decode json body", http.StatusBadRequest)
		return
	}

	validator := validator.New()
	if err := validator.Struct(question); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.questions.Add(question)
	if err != nil {
		log.Println("Internal error adding question", err)
		http.Error(w, "Internal error adding question", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(question)
	if err != nil {
		log.Println("err encoding json response add question", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s Server) updateQuestion(w http.ResponseWriter, r *http.Request) {

	var question domain.Question
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Incorrect media type", http.StatusUnsupportedMediaType)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&question)
	if err != nil {
		http.Error(w, "failed to decode json body", http.StatusBadRequest)
		return
	}

	validator := validator.New()
	if err := validator.Struct(question); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.questions.Update(question)

	if errors.Is(err, domain.ErrNoQuestionFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		log.Println("Internal error updating question", err)
		http.Error(w, "Internal error updating question", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(question)
	if err != nil {
		log.Println("err encoding json response update question", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
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
