package main

import (
	"github.com/coalaura/plain"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

var (
	manager = NewWidgetManager()
	log     = plain.New(plain.WithDate(plain.RFC3339Local))
)

func main() {
	manager.RegisterDefault()

	app := fiber.New(fiber.Config{
		ProxyHeader: fiber.HeaderXForwardedFor,
		Views:       html.New("./templates", ".html"),
	})

	app.Use(recover.New())
	app.Use(log.Middleware())

	app.Static("/", "./static")

	app.Get("/widgets.json", func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-Type", "application/json")

		return c.Send(manager.JSON())
	})

	app.Get("/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")

		return manager.Render(c, name)
	})

	log.Println("Listening on http://localhost:4777/")
	log.MustFail(app.Listen(":4777"))
}
