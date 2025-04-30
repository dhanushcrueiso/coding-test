package main

import (
	"github.com/dhanushcrueiso/coding-test/src/router"

	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New(fiber.Config{
		AppName: "Acronis-DataStore",
	})

	//server := handlers.NewServer()
	router.MountRoutes(app)
	app.Listen(":3000")
}
