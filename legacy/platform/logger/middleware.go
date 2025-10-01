package logger

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	SessionIDKey ContextKey = "session_id"
	LoggerKey    ContextKey = "logger"
)

func FiberMiddleware(logger *Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		sessionID := c.Params("sessionId")
		if sessionID == "" {
			sessionID = c.Params("session_id")
		}

		requestLogger := logger.WithRequest(requestID)
		if sessionID != "" {
			requestLogger = requestLogger.WithSession(sessionID)
		}

		c.Locals(string(LoggerKey), requestLogger)
		c.Locals(string(RequestIDKey), requestID)
		if sessionID != "" {
			c.Locals(string(SessionIDKey), sessionID)
		}

		requestLogger.EventDebug("http.request.start").
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("user_agent", c.Get("User-Agent")).
			Str("remote_ip", c.IP()).
			Msg("")

		err := c.Next()

		elapsed := time.Since(start)
		event := requestLogger.Event("http.request.complete").
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Int64("elapsed_ms", elapsed.Milliseconds()).
			Int("response_size", len(c.Response().Body()))

		if err != nil {
			event = event.Err(err)
		}

		event.Msg("")

		return err
	}
}

func FromFiberContext(c *fiber.Ctx) *Logger {
	if logger, ok := c.Locals(string(LoggerKey)).(*Logger); ok {
		return logger
	}
	return New()
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	return New()
}

func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals(string(RequestIDKey)).(string); ok {
		return requestID
	}
	return ""
}

func GetSessionID(c *fiber.Ctx) string {
	if sessionID, ok := c.Locals(string(SessionIDKey)).(string); ok {
		return sessionID
	}
	return ""
}
