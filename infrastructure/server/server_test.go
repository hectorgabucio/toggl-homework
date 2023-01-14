package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/togglhire/backend-homework/infrastructure/sql"
	"github.com/togglhire/backend-homework/usecase"
)

func TestServer_handleStatus(t *testing.T) {
	db := sql.SetupSQLConnection("test.db")
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
