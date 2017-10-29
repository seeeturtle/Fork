package main

import (
	"github.com/joshua1b/Fork/app"
	"github.com/joshua1b/Fork/config"
)

func main() {
	config := config.GetConfig()

	app := &app.App{}
	app.Initialize(config)
	app.Run(":3000")
}
