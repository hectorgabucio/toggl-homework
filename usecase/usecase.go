package usecase

import (
	"fmt"
	"log"

	"github.com/togglhire/backend-homework/domain"
)

type Questions struct {
	repo domain.QuestionRepository
}

func NewQuestions(questionRepository domain.QuestionRepository) Questions {
	return Questions{repo: questionRepository}
}

func (q Questions) GetAll() []domain.Question {
	questions, err := q.repo.GetAll()
	if err != nil {
		log.Printf("err getting all questions: %s", err)
		return []domain.Question{}
	}

	return questions

}

func (q Questions) Add(question domain.Question) error {
	err := q.repo.Add(question)
	if err != nil {
		return fmt.Errorf("err adding question:%w", err)
	}
	return nil
}

func (q Questions) Update(question domain.Question) error {
	err := q.repo.Update(question)
	if err != nil {
		return fmt.Errorf("err updating question:%w", err)
	}
	return nil
}
