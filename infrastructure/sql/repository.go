package sql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/togglhire/backend-homework/domain"
)

type Repository struct {
	db *sqlx.DB
}

type Option struct {
	ID         int    `db:"id"`
	Body       string `db:"body"`
	Correct    bool   `db:"correct"`
	QuestionID int    `db:"question_id"`
}

type Question struct {
	ID      int    `db:"id"`
	Body    string `db:"body"`
	Options []Option
}

func NewRepo(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r Repository) GetAll() ([]domain.Question, error) {
	var rows []Question
	query := `select q.* from question q order by q.id desc;`
	if err := r.db.Select(&rows, query); err != nil {
		return nil, fmt.Errorf("err query get all questions:%w", err)
	}

	// TODO sqlx doesnt support good inner join binding...

	for i := range rows {
		err := r.db.Select(&rows[i].Options, "SELECT option.id, option.body, option.correct FROM option WHERE question_id = ? order by option.id asc;", rows[i].ID)
		if err != nil {
			return nil, fmt.Errorf("err query get options of questions:%w", err)
		}
	}

	models := make([]domain.Question, len(rows))
	for i, row := range rows {
		models[i] = convertToDomain(row)
	}
	return models, nil
}

func (r Repository) Add(question domain.Question) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("err opening trx for adding question:%w", err)
	}

	_, err = tx.Exec("INSERT INTO question (id, body) VALUES (?, ?)", question.ID, question.Body)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec adding question:%w", err)
	}

	var optionsToAdd []Option
	for _, opt := range question.Options {
		optionsToAdd = append(optionsToAdd, Option{Body: opt.Body, Correct: opt.Correct, QuestionID: question.ID})
	}

	_, err = tx.NamedExec(`INSERT INTO option (body, correct, question_id)
	VALUES (:body, :correct,:question_id)`, optionsToAdd)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec adding options:%w", err)

	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err commit trx add question:%w", err)
	}

	return nil
}

func (r Repository) Update(question domain.Question) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("err opening trx for adding question:%w", err)
	}

	row := tx.QueryRow("select 1 from question where question.id = ?", question.ID)
	err = row.Err()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err checking if question exists:%w", err)
	}
	var exist int
	err = row.Scan(&exist)
	if err != nil {
		_ = tx.Rollback()
		return domain.ErrNoQuestionFound
	}
	if exist != 1 {
		_ = tx.Rollback()
		return fmt.Errorf("err question with that id doesnt exist:%w", err)
	}

	_, err = tx.NamedExec(`UPDATE question SET body=:body WHERE id = :id`, question)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec updating question:%w", err)
	}

	_, err = tx.Exec(`DELETE FROM option where question_id = ?`, question.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec deleting option before creating:%w", err)
	}

	var optionsToAdd []Option
	for _, opt := range question.Options {
		optionsToAdd = append(optionsToAdd, Option{Body: opt.Body, Correct: opt.Correct, QuestionID: question.ID})
	}

	_, err = tx.NamedExec(`INSERT INTO option (body, correct, question_id)
		VALUES (:body, :correct,:question_id)`, optionsToAdd)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec adding options on update:%w", err)

	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err commit trx update question:%w", err)
	}

	return nil
}

func convertToDomain(modelVoice Question) domain.Question {
	options := make([]domain.Option, 0)
	for _, opt := range modelVoice.Options {
		options = append(options, domain.Option{Body: opt.Body, Correct: opt.Correct})
	}
	return domain.Question{
		ID:      modelVoice.ID,
		Body:    modelVoice.Body,
		Options: options,
	}
}
