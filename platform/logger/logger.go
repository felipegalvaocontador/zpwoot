package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"zpwoot/platform/config"
)


type Logger struct {
	logger zerolog.Logger
	config config.LogConfig
}


func New(cfg config.LogConfig) *Logger {
	return NewWithConfig(cfg)
}


func NewWithConfig(cfg config.LogConfig) *Logger {

	cfg = validateLogConfig(cfg)


	logLevel := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(logLevel)


	zerolog.TimeFieldFormat = time.RFC3339


	var writer io.Writer = os.Stdout
	if cfg.Output == "stderr" {
		writer = os.Stderr
	}


	if cfg.Format == "console" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}


		if cfg.Caller {
			consoleWriter.FormatCaller = func(i interface{}) string {
				if caller, ok := i.(string); ok {
					return formatCaller(caller)
				}
				return ""
			}
		}

		writer = consoleWriter
	}


	ctx := zerolog.New(writer).With().
		Timestamp()


	if cfg.Caller {
		ctx = ctx.CallerWithSkipFrameCount(3)
	}

	logger := ctx.Logger()

	return &Logger{
		logger: logger,
		config: cfg,
	}
}


func NewFromAppConfig(appConfig *config.Config) *Logger {
	return New(appConfig.Log)
}


func (l *Logger) WithModule(module string) *Logger {
	newLogger := l.logger.With().Str("component", module).Logger()
	return &Logger{
		logger: newLogger,
		config: l.config,
	}
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


func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}


func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
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


func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		logger: ctx.Logger(),
		config: l.config,
	}
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




func (l *Logger) Event(event string) *zerolog.Event {
	return l.logger.Info().Str("event", event)
}


func (l *Logger) EventDebug(event string) *zerolog.Event {
	return l.logger.Debug().Str("event", event)
}


func (l *Logger) EventWarn(event string) *zerolog.Event {
	return l.logger.Warn().Str("event", event)
}


func (l *Logger) EventError(event string) *zerolog.Event {
	return l.logger.Error().Str("event", event)
}




func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}


func (l *Logger) GetConfig() config.LogConfig {
	return l.config
}


func (l *Logger) IsDebugEnabled() bool {
	return l.logger.GetLevel() <= zerolog.DebugLevel
}


func (l *Logger) IsTraceEnabled() bool {
	return l.logger.GetLevel() <= zerolog.TraceLevel
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


func validateLogConfig(cfg config.LogConfig) config.LogConfig {

	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true, "disabled": true,
	}
	if !validLevels[strings.ToLower(cfg.Level)] {
		cfg.Level = "info"
	}


	if cfg.Format != "console" && cfg.Format != "json" {
		cfg.Format = "json"
	}


	if cfg.Output != "stdout" && cfg.Output != "stderr" && cfg.Output != "file" {
		cfg.Output = "stdout"
	}

	return cfg
}


func formatCaller(caller string) string {

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




func DevelopmentConfig() config.LogConfig {
	return config.LogConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
		Caller: true,
	}
}


func ProductionConfig() config.LogConfig {
	return config.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
		Caller: false,
	}
}


func TestConfig() config.LogConfig {
	return config.LogConfig{
		Level:  "warn",
		Format: "json",
		Output: "stdout",
		Caller: false,
	}
}
