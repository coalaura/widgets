package main

import "github.com/gofiber/fiber/v2"

func abort(c *fiber.Ctx, code int) error {
	return c.SendStatus(code)
}
