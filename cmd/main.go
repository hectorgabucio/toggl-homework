package main

import (
	"log"

	"github.com/togglhire/backend-homework/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal("app closed, reason: ", err)
	}
}
