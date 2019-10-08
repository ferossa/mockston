package main

import (
	"github.com/ferossa/mockston/internal/app"
)

func main() {
	m := app.NewMockston()
	m.Run()
}
