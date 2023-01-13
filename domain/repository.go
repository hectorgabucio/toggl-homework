package domain

import "github.com/togglhire/backend-homework/model"

type QuestionRepository interface {
	GetAll() ([]model.Question, error)
}
