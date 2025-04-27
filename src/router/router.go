package router

import (
	"CODING-TEST/src/handlers"

	"github.com/gofiber/fiber/v2"
)

func MountRoutes(app *fiber.App) {
	apiGroup := app.Group("/api")
	controller := handlers.Handler{}
	apiGroup.Get("/health", controller.GetHealth)
	stringsGroup := apiGroup.Group("/strings")
	{
		stringsGroup.Post("/:key", controller.SetStringData)
		stringsGroup.Get("/:key", controller.GetStringData)
		stringsGroup.Put("/:key", controller.UpdateStringData)
		stringsGroup.Delete("/:key", controller.DeleteStringData)
	}
	TtlGroup := apiGroup.Group("/ttl")
	{
		TtlGroup.Get("/:key", controller.GetTtlData)
		TtlGroup.Post("/:key", controller.SetTtlData)
	}
	ListGroup := apiGroup.Group("/list")
	{
		ListGroup.Get("/:key", controller.GetListData)
		ListGroup.Post("/:key", controller.SetListData)
		ListGroup.Delete("/:key", controller.DeleteListData)
		ListGroup.Patch("/:key/:operation", controller.UpdateListData)
	}

}
