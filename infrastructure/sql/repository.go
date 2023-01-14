package sql

import (
	"fmt"
	"sort"

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

type OrderedOptions []Option

func (a OrderedOptions) Len() int { return len(a) }
func (a OrderedOptions) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}
func (a OrderedOptions) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (Option) TableName() string {
	return "option"
}

type Question struct {
	ID      int    `db:"id"`
	Body    string `db:"body"`
	Options []Option
}

type OrderedQuestions []Question

func (a OrderedQuestions) Len() int { return len(a) }
func (a OrderedQuestions) Less(i, j int) bool {
	return a[i].ID > a[j].ID
}
func (a OrderedQuestions) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (Question) TableName() string {
	return "question"
}

func NewRepo(db *gorm.DB) Repository {
	return Repository{db: db}
}

func (r Repository) GetAll() ([]domain.Question, error) {
	var rows []Question
	err := r.db.Preload("Options").Find(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("err query get all questions:%w", err)
	}

	return convertToDomain(rows), nil
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

func convertToDomain(questions []Question) []domain.Question {

	var orderQuestions OrderedQuestions = questions
	sort.Sort(orderQuestions)
	domainQuestions := make([]domain.Question, 0)

	for _, question := range orderQuestions {
		options := make([]domain.Option, 0)

		var orderOpt OrderedOptions = question.Options
		sort.Sort(orderOpt)
		for _, opt := range orderOpt {
			options = append(options, domain.Option{Body: opt.Body, Correct: opt.Correct})
		}

		domainQuestion := domain.Question{
			ID:      question.ID,
			Body:    question.Body,
			Options: options,
		}
		domainQuestions = append(domainQuestions, domainQuestion)
	}

	return domainQuestions

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
