package middleware

import (
	"zpwoot/internal/app"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

func Metrics(container *app.Container, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		container.GetCommonUseCase().IncrementRequestCount()

		err := c.Next()

		if err != nil {
			container.GetCommonUseCase().IncrementErrorCount()

			logger.ErrorWithFields("Request error", map[string]interface{}{
				"method":     c.Method(),
				"path":       c.Path(),
				"status":     c.Response().StatusCode(),
				"error":      err.Error(),
				"request_id": c.Locals("request_id"),
			})
		}

		return err
	}
}
