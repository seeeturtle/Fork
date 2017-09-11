package main

import (
	"github.com/joshua1b/SchoolMeal/app"
	"github.com/joshua1b/SchoolMeal/config"
)

func main() {
	config := config.GetConfig()

	app := &app.App{}
	app.Initialize(config)
	app.Run(":3000")
}
