package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

func APIKeyAuth(cfg *config.Config, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/swagger") || strings.Contains(path, "/chatwoot/webhook") {
			return c.Next()
		}

		apiKey := c.Get("Authorization")
		if apiKey == "" {
			apiKey = c.Get("X-API-Key")
		}

		if apiKey == "" {
			logger.WarnWithFields("Missing API key", map[string]interface{}{
				"path":   path,
				"method": c.Method(),
				"ip":     c.IP(),
			})
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "API key is required. Provide it via Authorization header or X-API-Key header",
				"code":    "MISSING_API_KEY",
			})
		}

		if apiKey != cfg.GlobalAPIKey {
			logger.WarnWithFields("Invalid API key", map[string]interface{}{
				"path":    path,
				"method":  c.Method(),
				"ip":      c.IP(),
				"api_key": maskAPIKey(apiKey),
			})
			return c.Status(401).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid API key",
				"code":    "INVALID_API_KEY",
			})
		}

		logger.DebugWithFields("API key authenticated", map[string]interface{}{
			"path":    path,
			"method":  c.Method(),
			"ip":      c.IP(),
			"api_key": maskAPIKey(apiKey),
		})

		c.Locals("api_key", apiKey)
		c.Locals("authenticated", true)

		return c.Next()
	}
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 12 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:8] + strings.Repeat("*", len(apiKey)-12) + apiKey[len(apiKey)-4:]
}

func GetAPIKeyFromContext(c *fiber.Ctx) string {
	if apiKey, ok := c.Locals("api_key").(string); ok {
		return apiKey
	}
	return ""
}

func IsAuthenticated(c *fiber.Ctx) bool {
	if authenticated, ok := c.Locals("authenticated").(bool); ok {
		return authenticated
	}
	return false
}
