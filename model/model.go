package model

// TODO separate domain from API json tags

type Question struct {
	Body    string
	Options []Option
}

type Option struct {
	Body    string
	Correct bool
}
