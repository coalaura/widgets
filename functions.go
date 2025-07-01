package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func abort(c *fiber.Ctx, code int) error {
	return c.SendStatus(code)
}

func slice(values ...any) []string {
	result := make([]string, len(values))

	for i, value := range values {
		if str, ok := value.(string); ok {
			result[i] = str
		} else {
			result[i] = fmt.Sprint(value)
		}
	}

	return result
}
