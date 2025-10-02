package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
	config *LogConfig
}

func New() *Logger {
	env := strings.ToLower(os.Getenv("ZPWOOT_ENV"))
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var config *LogConfig
	switch env {
	case "development", "dev":
		config = DevelopmentConfig()
	case "production", "prod":
		config = ProductionConfig()
	default:
		config = DevelopmentConfig()
	}

	if logLevel != "" {
		config.Level = logLevel
	}

	return NewWithConfig(config)
}

func NewWithConfig(config *LogConfig) *Logger {
	config.Validate()

	logLevel := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(logLevel)

	zerolog.TimeFieldFormat = time.RFC3339

	var writer io.Writer = os.Stdout

	if config.Output == "file" {
		writer = os.Stdout
	}

	if config.Format == "console" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}

		if config.Caller {
			consoleWriter.FormatCaller = func(i interface{}) string {
				if caller, ok := i.(string); ok {
					if strings.Contains(caller, "/workspaces/zpwoot/") {
						relativePath := strings.TrimPrefix(caller, "/workspaces/zpwoot/")
						return relativePath
					}
					if strings.Contains(caller, "zpwoot/") {
						parts := strings.Split(caller, "zpwoot/")
						if len(parts) > 1 {
							return parts[len(parts)-1]
						}
					}
					return filepath.Base(caller)
				}
				return ""
			}
		}

		writer = consoleWriter
	}

	ctx := zerolog.New(writer).With().
		Timestamp().
		Str("service", "zpwoot")

	if env := os.Getenv("ZPWOOT_ENV"); env != "" {
		ctx = ctx.Str("env", env)
	}

	if config.Caller {
		ctx = ctx.CallerWithSkipFrameCount(3)
	}

	logger := ctx.Logger()

	return &Logger{
		logger: logger,
		config: config,
	}
}

func (l *Logger) Event(event string) *zerolog.Event {
	return l.logger.Info().Str("event", event)
}

func (l *Logger) EventDebug(event string) *zerolog.Event {
	return l.logger.Debug().Str("event", event)
}

func (l *Logger) EventError(event string) *zerolog.Event {
	return l.logger.Error().Str("event", event)
}

func (l *Logger) EventWarn(event string) *zerolog.Event {
	return l.logger.Warn().Str("event", event)
}

func (l *Logger) WithSession(sessionID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("session_id", sessionID).Logger(),
		config: l.config,
	}
}

func (l *Logger) WithRequest(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
		config: l.config,
	}
}

func (l *Logger) WithMessage(messageID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("message_id", messageID).Logger(),
		config: l.config,
	}
}

func (l *Logger) WithElapsed(start time.Time) *Logger {
	elapsed := time.Since(start).Milliseconds()
	return &Logger{
		logger: l.logger.With().Int64("elapsed_ms", elapsed).Logger(),
		config: l.config,
	}
}

func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info", "":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Error()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
		config: l.config,
	}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
		config: l.config,
	}
}

func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}
