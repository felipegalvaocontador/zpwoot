package middleware

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"

	"zpwoot/platform/logger"
)

type LoggerConfig struct {
	Output io.Writer
	Logger *logger.Logger
	Format string
}

func NewLogger(customLogger *logger.Logger) fiber.Handler {
	return NewLoggerWithConfig(LoggerConfig{
		Logger: customLogger,
	})
}

func NewLoggerWithConfig(config LoggerConfig) fiber.Handler {
	if config.Logger == nil {
		config.Logger = logger.New()
	}

	if config.Format == "" {
		config.Format = "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n"
	}

	if config.Output == nil {
		config.Output = &httpLogWriter{logger: config.Logger}
	}

	return fiberLogger.New(fiberLogger.Config{
		Format:     config.Format,
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     config.Output,
		CustomTags: map[string]fiberLogger.LogFunc{
			"custom_log": func(output fiberLogger.Buffer, c *fiber.Ctx, data *fiberLogger.Data, extraParam string) (int, error) {
				logHTTPRequest(config.Logger, c, data)
				return 0, nil
			},
		},
	})
}

type httpLogWriter struct {
	logger *logger.Logger
}

func (w *httpLogWriter) Write(p []byte) (int, error) {
	logLine := strings.TrimSpace(string(p))
	if logLine == "" {
		return len(p), nil
	}

	parts := strings.Split(logLine, " | ")
	if len(parts) >= 6 {
		timestamp := parts[0]
		status := parts[1]
		latency := parts[2]
		ip := parts[3]
		method := parts[4]
		path := parts[5]
		errorMsg := ""
		if len(parts) > 6 {
			errorMsg = parts[6]
		}

		statusCode, err := strconv.Atoi(status)
		if err != nil {
			statusCode = 0 // Default to 0 if conversion fails
		}

		fields := map[string]interface{}{
			"component":   "http",
			"timestamp":   timestamp,
			"status_code": statusCode,
			"latency":     latency,
			"ip":          ip,
			"method":      method,
			"path":        path,
		}

		if errorMsg != "" && errorMsg != "-" {
			fields["error"] = errorMsg
		}

		message := fmt.Sprintf("%s %s", method, path)

		switch {
		case statusCode >= 500:
			w.logger.ErrorWithFields(message, fields)
		case statusCode >= 400:
			w.logger.WarnWithFields(message, fields)
		default:
			w.logger.InfoWithFields(message, fields)
		}
	} else {
		w.logger.Info(logLine)
	}

	return len(p), nil
}

func logHTTPRequest(logger *logger.Logger, c *fiber.Ctx, data *fiberLogger.Data) {
	statusCode := c.Response().StatusCode()

	fields := map[string]interface{}{
		"component":      "http",
		"method":         c.Method(),
		"path":           c.Path(),
		"route":          c.Route().Path,
		"status_code":    statusCode,
		"latency_ms":     data.Stop.Sub(data.Start).Milliseconds(),
		"ip":             c.IP(),
		"user_agent":     c.Get("User-Agent"),
		"content_length": len(c.Response().Body()),
	}

	if c.Request().URI().QueryString() != nil {
		fields["query"] = string(c.Request().URI().QueryString())
	}

	if requestID := c.Get("X-Request-ID"); requestID != "" {
		fields["request_id"] = requestID
	}

	if sessionID := c.Get("X-Session-ID"); sessionID != "" {
		fields["session_id"] = sessionID
	} else if sessionID := c.Query("session_id"); sessionID != "" {
		fields["session_id"] = sessionID
	}

	if err := c.Locals("error"); err != nil {
		fields["error"] = fmt.Sprintf("%v", err)
	}

	message := fmt.Sprintf("%s %s", c.Method(), c.Path())

	switch {
	case statusCode >= 500:
		logger.ErrorWithFields(message, fields)
	case statusCode >= 400:
		logger.WarnWithFields(message, fields)
	case statusCode >= 300:
		logger.InfoWithFields(message, fields)
	default:
		logger.InfoWithFields(message, fields)
	}
}

func HTTPLogger(logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		statusCode := c.Response().StatusCode()

		fields := map[string]interface{}{
			"component":      "http",
			"method":         c.Method(),
			"path":           c.Path(),
			"route":          c.Route().Path,
			"status_code":    statusCode,
			"latency_ms":     latency.Milliseconds(),
			"latency_human":  latency.String(),
			"ip":             c.IP(),
			"user_agent":     c.Get("User-Agent"),
			"content_length": len(c.Response().Body()),
			"protocol":       c.Protocol(),
		}

		if queryString := string(c.Request().URI().QueryString()); queryString != "" {
			fields["query"] = queryString
		}

		if contentType := c.Get("Content-Type"); contentType != "" {
			fields["content_type"] = contentType
		}

		if sessionID := c.Get("X-Session-ID"); sessionID != "" {
			fields["session_id"] = sessionID
		}

		if requestID := c.Get("X-Request-ID"); requestID != "" {
			fields["request_id"] = requestID
		}

		if err != nil {
			fields["error"] = err.Error()
		}

		message := fmt.Sprintf("HTTP %s %s", c.Method(), c.Path())

		switch {
		case err != nil:
			logger.ErrorWithFields(message, fields)
		case statusCode >= 500:
			logger.ErrorWithFields(message, fields)
		case statusCode >= 400:
			logger.WarnWithFields(message, fields)
		case statusCode >= 300:
			logger.InfoWithFields(message, fields)
		default:
			logger.DebugWithFields(message, fields)
		}

		return err
	}
}
