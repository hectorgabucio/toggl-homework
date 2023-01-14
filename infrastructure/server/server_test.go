package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/togglhire/backend-homework/domain"
	"github.com/togglhire/backend-homework/infrastructure/sql"
	"github.com/togglhire/backend-homework/usecase"
)

func buildBufJson(question domain.Question, t *testing.T) *bytes.Buffer {
	input, err := json.Marshal(question)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewBuffer(input)
}

func TestServer_handleStatus(t *testing.T) {
	db := sql.SetupSQLConnection("test.db")
	defer os.Remove("test.db")
	_, srv := NewServer(context.Background(), 0, usecase.NewQuestions(sql.NewRepo(db)))

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		expectedStatus int
	}{
		{name: "Get status should return OK", args: args{r: httptest.NewRequest(http.MethodGet, "/status", nil)}, expectedStatus: http.StatusOK},
		{name: "Status with invalid method should not be OK", args: args{r: httptest.NewRequest(http.MethodPost, "/status", nil)}, expectedStatus: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			srv.handleStatus(rr, tt.args.r)
			if rr.Result().StatusCode != tt.expectedStatus {
				t.Errorf("Status code returned, %d, did not match expected code %d", rr.Result().StatusCode, tt.expectedStatus)
			}
		})
	}
}

func TestServer_ListQuestionsOrderShouldBeStable(t *testing.T) {
	db := sql.SetupSQLConnection("test.db")
	defer os.Remove("test.db")
	repo := sql.NewRepo(db)

	_ = repo.Add(domain.Question{ID: 1, Body: "one", Options: []domain.Option{
		{Body: "option one", Correct: false},
	}})

	_ = repo.Add(domain.Question{ID: 3, Body: "three", Options: []domain.Option{
		{Body: "option one for question 3", Correct: false},
	}})

	_ = repo.Add(domain.Question{ID: 2, Body: "two", Options: []domain.Option{
		{Body: "option one for question 2", Correct: false},
	}})

	_, srv := NewServer(context.Background(), 0, usecase.NewQuestions(repo))
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/questions", nil)
	srv.listQuestions(rr, r)
	if rr.Result().StatusCode != http.StatusOK {
		t.Errorf("Status code returned, %d, did not match expected code %d", rr.Result().StatusCode, http.StatusOK)
	}

	var response []domain.Question
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("json response should be unmarshable")
	}
	expected := `[{"id":3,"body":"three","options":[{"body":"option one for question 3","correct":false}]},{"id":2,"body":"two","options":[{"body":"option one for question 2","correct":false}]},{"id":1,"body":"one","options":[{"body":"option one","correct":false}]}]`
	got := strings.TrimSpace(rr.Body.String())
	eq := strings.Compare(got, expected)

	if eq != 0 {
		t.Errorf("json returned, %s, did not match expected json %s", got, expected)
	}

	rr = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/questions", nil)
	srv.listQuestions(rr, r)
	got = strings.TrimSpace(rr.Body.String())
	eq = strings.Compare(got, expected)

	if eq != 0 {
		t.Errorf("json returned, %s, did not match expected json %s", got, expected)
	}

}

func TestServer_addQuestion(t *testing.T) {
	db := sql.SetupSQLConnection("test.db")
	defer os.Remove("test.db")
	_, srv := NewServer(context.Background(), 0, usecase.NewQuestions(sql.NewRepo(db)))

	emptyOptQuestion := domain.Question{ID: 1, Body: "hello", Options: []domain.Option{}}
	validQuestion := domain.Question{ID: 1, Body: "hello",
		Options: []domain.Option{
			{Body: "option a"}, {Body: "option b", Correct: true},
		}}

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		expectedStatus int
	}{
		{name: "NON body request should fail with 400",
			args:           args{r: httptest.NewRequest(http.MethodPost, "/questions", nil)},
			expectedStatus: http.StatusBadRequest},
		{name: "invalid json request should fail with 400",
			args: args{r: httptest.NewRequest(http.MethodPost, "/questions",
				buildBufJson(emptyOptQuestion, t))},
			expectedStatus: http.StatusBadRequest},
		{name: "valid json request should be OK",
			args: args{r: httptest.NewRequest(http.MethodPost, "/questions",
				buildBufJson(validQuestion, t))},
			expectedStatus: http.StatusOK},
		{name: "create same question id should fail with 500",
			args: args{r: httptest.NewRequest(http.MethodPost, "/questions",
				buildBufJson(validQuestion, t))},
			expectedStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tt.args.r.Header.Add("Content-Type", "application/json")
			srv.addQuestion(rr, tt.args.r)
			if rr.Result().StatusCode != tt.expectedStatus {
				t.Errorf("Status code returned, %d, did not match expected code %d", rr.Result().StatusCode, tt.expectedStatus)
			}
		})
	}
}

func TestServer_updateQuestion(t *testing.T) {
	db := sql.SetupSQLConnection("test.db")
	defer os.Remove("test.db")
	repo := sql.NewRepo(db)
	_, srv := NewServer(context.Background(), 0, usecase.NewQuestions(repo))

	emptyOptQuestion := domain.Question{ID: 1, Body: "hello", Options: []domain.Option{}}
	validQuestionNonExistent := domain.Question{ID: 1, Body: "hello",
		Options: []domain.Option{
			{Body: "option a"}, {Body: "option b", Correct: true},
		}}

	validQuestionExistent := domain.Question{ID: 2, Body: "hello",
		Options: []domain.Option{
			{Body: "option a"}, {Body: "option b", Correct: true},
		}}
	_ = repo.Add(validQuestionExistent)

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		expectedStatus int
	}{
		{name: "NON body request should fail with 400",
			args:           args{r: httptest.NewRequest(http.MethodPut, "/questions", nil)},
			expectedStatus: http.StatusBadRequest},
		{name: "invalid json request should fail with 400",
			args: args{r: httptest.NewRequest(http.MethodPost, "/questions",
				buildBufJson(emptyOptQuestion, t))},
			expectedStatus: http.StatusBadRequest},
		{name: "valid question but not existent should throw 404",
			args: args{r: httptest.NewRequest(http.MethodPost, "/questions",
				buildBufJson(validQuestionNonExistent, t))},
			expectedStatus: http.StatusNotFound},
		{name: "valid question existent should give 200",
			args: args{r: httptest.NewRequest(http.MethodPost, "/questions",
				buildBufJson(validQuestionExistent, t))},
			expectedStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tt.args.r.Header.Add("Content-Type", "application/json")
			srv.updateQuestion(rr, tt.args.r)
			if rr.Result().StatusCode != tt.expectedStatus {
				t.Errorf("Status code returned, %d, did not match expected code %d", rr.Result().StatusCode, tt.expectedStatus)
			}
		})
	}
}
