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

// Logger wrapper para zerolog com funcionalidades específicas do zpwoot
type Logger struct {
	logger zerolog.Logger
	config config.LogConfig
}

// New cria um novo logger baseado na configuração
func New(cfg config.LogConfig) *Logger {
	return NewWithConfig(cfg)
}

// NewWithConfig cria logger com configuração específica
func NewWithConfig(cfg config.LogConfig) *Logger {
	// Validar configuração
	cfg = validateLogConfig(cfg)

	// Configurar nível global
	logLevel := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(logLevel)

	// Configurar formato de timestamp
	zerolog.TimeFieldFormat = time.RFC3339

	// Configurar writer de saída
	var writer io.Writer = os.Stdout
	if cfg.Output == "stderr" {
		writer = os.Stderr
	}

	// Configurar formato de saída
	if cfg.Format == "console" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}

		// Configurar formatação do caller se habilitado
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

	// Criar contexto base do logger
	ctx := zerolog.New(writer).With().
		Timestamp()

	// Adicionar caller se habilitado
	if cfg.Caller {
		ctx = ctx.CallerWithSkipFrameCount(3)
	}

	logger := ctx.Logger()

	return &Logger{
		logger: logger,
		config: cfg,
	}
}

// NewFromAppConfig cria logger a partir da configuração da aplicação
func NewFromAppConfig(appConfig *config.Config) *Logger {
	return New(appConfig.Log)
}

// WithModule cria um novo logger com módulo específico
func (l *Logger) WithModule(module string) *Logger {
	newLogger := l.logger.With().Str("component", module).Logger()
	return &Logger{
		logger: newLogger,
		config: l.config,
	}
}

// ===== MÉTODOS DE LOGGING =====

// Debug registra mensagem de debug
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf registra mensagem de debug formatada
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// DebugWithFields registra mensagem de debug com campos
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info registra mensagem informativa
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof registra mensagem informativa formatada
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// InfoWithFields registra mensagem informativa com campos
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn registra mensagem de aviso
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf registra mensagem de aviso formatada
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// WarnWithFields registra mensagem de aviso com campos
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error registra mensagem de erro
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf registra mensagem de erro formatada
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithFields registra mensagem de erro com campos
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	event := l.logger.Error()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal registra mensagem fatal e termina aplicação
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf registra mensagem fatal formatada e termina aplicação
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// ===== MÉTODOS DE CONTEXTO =====

// WithError adiciona erro ao contexto
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
		config: l.config,
	}
}

// WithField adiciona campo ao contexto
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
		config: l.config,
	}
}

// WithFields adiciona múltiplos campos ao contexto
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

// WithSession adiciona ID da sessão ao contexto
func (l *Logger) WithSession(sessionID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("session_id", sessionID).Logger(),
		config: l.config,
	}
}

// WithRequest adiciona ID da requisição ao contexto
func (l *Logger) WithRequest(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
		config: l.config,
	}
}

// WithMessage adiciona ID da mensagem ao contexto
func (l *Logger) WithMessage(messageID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("message_id", messageID).Logger(),
		config: l.config,
	}
}

// WithElapsed adiciona tempo decorrido ao contexto
func (l *Logger) WithElapsed(start time.Time) *Logger {
	elapsed := time.Since(start).Milliseconds()
	return &Logger{
		logger: l.logger.With().Int64("elapsed_ms", elapsed).Logger(),
		config: l.config,
	}
}

// ===== MÉTODOS DE EVENTO =====

// Event cria evento informativo
func (l *Logger) Event(event string) *zerolog.Event {
	return l.logger.Info().Str("event", event)
}

// EventDebug cria evento de debug
func (l *Logger) EventDebug(event string) *zerolog.Event {
	return l.logger.Debug().Str("event", event)
}

// EventWarn cria evento de aviso
func (l *Logger) EventWarn(event string) *zerolog.Event {
	return l.logger.Warn().Str("event", event)
}

// EventError cria evento de erro
func (l *Logger) EventError(event string) *zerolog.Event {
	return l.logger.Error().Str("event", event)
}

// ===== MÉTODOS UTILITÁRIOS =====

// GetZerologLogger retorna logger zerolog subjacente
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}

// GetConfig retorna configuração do logger
func (l *Logger) GetConfig() config.LogConfig {
	return l.config
}

// IsDebugEnabled verifica se debug está habilitado
func (l *Logger) IsDebugEnabled() bool {
	return l.logger.GetLevel() <= zerolog.DebugLevel
}

// IsTraceEnabled verifica se trace está habilitado
func (l *Logger) IsTraceEnabled() bool {
	return l.logger.GetLevel() <= zerolog.TraceLevel
}

// ===== FUNÇÕES AUXILIARES =====

// parseLogLevel converte string para zerolog.Level
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

// validateLogConfig valida e corrige configuração de log
func validateLogConfig(cfg config.LogConfig) config.LogConfig {
	// Validar nível
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true, "panic": true, "disabled": true,
	}
	if !validLevels[strings.ToLower(cfg.Level)] {
		cfg.Level = "info"
	}

	// Validar formato
	if cfg.Format != "console" && cfg.Format != "json" {
		cfg.Format = "json"
	}

	// Validar saída
	if cfg.Output != "stdout" && cfg.Output != "stderr" && cfg.Output != "file" {
		cfg.Output = "stdout"
	}

	return cfg
}

// formatCaller formata informação do caller
func formatCaller(caller string) string {
	// Remover prefixo do workspace se presente
	if strings.Contains(caller, "/workspaces/zpwoot/") {
		relativePath := strings.TrimPrefix(caller, "/workspaces/zpwoot/")
		return relativePath
	}

	// Remover prefixo do módulo se presente
	if strings.Contains(caller, "zpwoot/") {
		parts := strings.Split(caller, "zpwoot/")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
	}

	// Retornar apenas o nome do arquivo
	return filepath.Base(caller)
}

// ===== CONFIGURAÇÕES PRÉ-DEFINIDAS =====

// DevelopmentConfig retorna configuração para desenvolvimento
func DevelopmentConfig() config.LogConfig {
	return config.LogConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
		Caller: true,
	}
}

// ProductionConfig retorna configuração para produção
func ProductionConfig() config.LogConfig {
	return config.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
		Caller: false,
	}
}

// TestConfig retorna configuração para testes
func TestConfig() config.LogConfig {
	return config.LogConfig{
		Level:  "warn",
		Format: "json",
		Output: "stdout",
		Caller: false,
	}
}
