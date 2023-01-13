package sqliterepo

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/togglhire/backend-homework/domain"
)

type Repository struct {
	db *sqlx.DB
}

type dbQuestionModel struct {
	ID   int    `db:"id"`
	Body string `db:"body"`
}

func New(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r Repository) GetAll() ([]domain.Question, error) {
	var rows []dbQuestionModel
	query := "select * from question q order by q.id desc;"
	if err := r.db.Select(&rows, query); err != nil {
		return nil, fmt.Errorf("err get all questions:%w", err)
	}
	models := make([]domain.Question, len(rows))
	for i, row := range rows {
		models[i] = convertToDomain(row)
	}
	return models, nil
}

func convertToDomain(modelVoice dbQuestionModel) domain.Question {
	return domain.Question{
		Body: modelVoice.Body,
	}
}
