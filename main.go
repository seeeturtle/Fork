package main

import (
	"os"

	"github.com/joshua1b/SchoolMeal/app"
	"github.com/joshua1b/SchoolMeal/config"
)

func main() {
	config := config.GetConfig()

	app := &app.App{}
	app.Initialize(config)
	port := os.Getenv("PORT")
	app.Run(":" + port)
}
