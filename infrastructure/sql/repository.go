package sql

import (
	"fmt"

	"github.com/togglhire/backend-homework/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type Tabler interface {
	TableName() string
}

type Option struct {
	ID         int    `db:"id"`
	Body       string `db:"body"`
	Correct    bool   `db:"correct"`
	QuestionID int    `db:"question_id"`
}

func (Option) TableName() string {
	return "option"
}

type Question struct {
	ID      int    `db:"id"`
	Body    string `db:"body"`
	Options []Option
}

func (Question) TableName() string {
	return "question"
}

func NewRepo(db *gorm.DB) Repository {
	return Repository{db: db}
}

func (r Repository) GetAll() ([]domain.Question, error) {
	var rows []Question

	err := r.db.Model(&Question{}).Preload("Options").Order("id desc").Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err query get all questions:%w", err)
	}

	questions := make([]domain.Question, 0)
	for _, row := range rows {
		questions = append(questions, convertToDomain(row))
	}

	return questions, nil
}

func (r Repository) Add(question domain.Question) error {
	tx := r.db.Begin()

	dbQuestion := convertToDBModel(question)

	if err := tx.Create(&dbQuestion).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("err sql exec adding question:%w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err commit trx add question:%w", err)
	}

	return nil
}

func (r Repository) Update(question domain.Question) error {

	tx := r.db.Begin()

	dbQuestion := convertToDBModel(question)

	var dbQuestionExists Question
	tx.First(&dbQuestionExists, question.ID)
	if dbQuestionExists.ID != question.ID {
		_ = tx.Rollback()
		return domain.ErrNoQuestionFound
	}

	err := tx.Model(&dbQuestion).Updates(dbQuestion).Error
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec updating question:%w", err)
	}

	err = tx.Exec(`DELETE FROM option where question_id = ?`, question.ID).Error
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec deleting option before creating:%w", err)
	}

	err = tx.Create(dbQuestion.Options).Error
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err sql exec adding options on update:%w", err)

	}

	err = tx.Commit().Error
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("err commit trx update question:%w", err)
	}

	return nil
}

func convertToDomain(question Question) domain.Question {
	options := make([]domain.Option, 0)
	for _, opt := range question.Options {
		options = append(options, domain.Option{Body: opt.Body, Correct: opt.Correct})
	}
	return domain.Question{
		ID:      question.ID,
		Body:    question.Body,
		Options: options,
	}
}
func convertToDBModel(question domain.Question) Question {
	dbOptions := make([]Option, 0)

	for _, opt := range question.Options {
		dbOptions = append(dbOptions, Option{Body: opt.Body, Correct: opt.Correct})
	}

	return Question{
		ID:      question.ID,
		Body:    question.Body,
		Options: dbOptions,
	}
}
