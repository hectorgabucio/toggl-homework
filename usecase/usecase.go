package usecase

import (
	"log"

	"github.com/togglhire/backend-homework/domain"
	"github.com/togglhire/backend-homework/model"
)

type Questions struct {
	repo domain.QuestionRepository
}

func NewQuestions(questionRepository domain.QuestionRepository) Questions {
	return Questions{repo: questionRepository}
}

func (q Questions) GetAll() []model.Question {
	questions, err := q.repo.GetAll()
	if err != nil {
		log.Printf("err getting all questions: %s", err)
		return []model.Question{}
	}

	return questions

}
