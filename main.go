package main

import (
	"github.com/joshua1b/Plate/app"
	"github.com/joshua1b/Plate/config"
)

func main() {
	config := config.GetConfig()

	app := &app.App{}
	app.Initialize(config)
	app.Run(":3000")
}
