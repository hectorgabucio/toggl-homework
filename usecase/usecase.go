package usecase

import "github.com/togglhire/backend-homework/model"

type Questions struct{}

func NewQuestions() Questions {
	return Questions{}
}

func (q Questions) GetAll() []model.Question {
	// TODO remove and use datasource
	questions := []model.Question{
		{Body: "hello", Options: []model.Option{{Body: "bye", Correct: true}}},
	}

	return questions

}
