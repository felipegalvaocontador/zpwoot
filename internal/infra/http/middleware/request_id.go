package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"zpwoot/platform/logger"
)

func RequestID(logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		c.Locals("request_id", requestID)

		requestLogger := logger.WithField("request_id", requestID)
		c.Locals("logger", requestLogger)

		return c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func GetLoggerFromContext(c *fiber.Ctx) *logger.Logger {
	if logger, ok := c.Locals("logger").(*logger.Logger); ok {
		return logger
	}
	return logger.New()
}

func LogError(c *fiber.Ctx, err error, message string) {
	requestLogger := GetLoggerFromContext(c)

	fields := map[string]interface{}{
		"component": "http",
		"method":    c.Method(),
		"path":      c.Path(),
		"ip":        c.IP(),
	}

	if requestID := c.Locals("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	requestLogger.ErrorWithFields(message, fields)
}

func LogInfo(c *fiber.Ctx, message string, additionalFields ...map[string]interface{}) {
	requestLogger := GetLoggerFromContext(c)

	fields := map[string]interface{}{
		"component": "http",
		"method":    c.Method(),
		"path":      c.Path(),
	}

	if requestID := c.Locals("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	for _, additional := range additionalFields {
		for k, v := range additional {
			fields[k] = v
		}
	}

	requestLogger.InfoWithFields(message, fields)
}
