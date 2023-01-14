package domain

import (
	"fmt"
)

var ErrNoQuestionFound = fmt.Errorf("not found question")

type Question struct {
	ID      int      `json:"id" validate:"required"`
	Body    string   `json:"body" validate:"required,min=1,max=255"`
	Options []Option `json:"options" validate:"required,min=2,max=10,dive"`
}

type Option struct {
	Body    string `json:"body" validate:"required,min=1,max=255"`
	Correct bool   `json:"correct"`
}

type QuestionRepository interface {
	GetAll() ([]Question, error)
	Add(Question) error
	Update(Question) error
}
