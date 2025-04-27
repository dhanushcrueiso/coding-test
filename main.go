package main

import (
	"github.com/dhanushcrueiso/coding-test/src/router"

	"github.com/dhanushcrueiso/coding-test/internal/db"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db.InitDb()

	app := fiber.New(fiber.Config{
		AppName: "Acronis-DataStore",
	})
	router.MountRoutes(app)
	app.Listen(":3000")
}
