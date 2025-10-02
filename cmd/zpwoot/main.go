// @title ZPWoot WhatsApp API
// @version 2.0.0
// @description A comprehensive WhatsApp Business API built with Go. Provides endpoints for session management, messaging, contacts, groups, media handling, and integrations with Chatwoot.
// @termsOfService https://github.com/zpwoot/zpwoot/blob/main/LICENSE

// @contact.name ZPWoot API Support
// @contact.url https://github.com/zpwoot/zpwoot
// @contact.email support@zpwoot.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description API Key authentication. Use: YOUR_API_KEY

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/services"
	"zpwoot/platform/config"
	"zpwoot/platform/container"
	"zpwoot/platform/database"
	"zpwoot/platform/logger"

	_ "zpwoot/docs/swagger" // Import docs for swagger
)

const (
	appName    = "zpwoot"
	appVersion = "2.0.0"
)

func main() {
	// Carregar configuração primeiro
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Exibir banner com configuração
	printBanner(cfg)

	// Criar contexto principal da aplicação
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicializar logger
	log := logger.NewFromAppConfig(cfg)
	log.InfoWithFields("Starting zpwoot application", map[string]interface{}{
		"version":     appVersion,
		"environment": cfg.Environment,
		"port":        cfg.Server.Port,
	})

	// Inicializar banco de dados
	log.Info("Initializing database connection...")
	db, err := database.NewFromAppConfig(cfg, log)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize database: %v", err))
	}
	defer db.Close()

	// Executar migrações se habilitado
	if cfg.Database.AutoMigrate {
		log.Info("Running database migrations...")
		if err := runMigrations(db, log); err != nil {
			log.Fatal(fmt.Sprintf("Failed to run migrations: %v", err))
		}
	}

	// Inicializar container de DI
	log.Info("Initializing dependency injection container...")
	containerConfig := &container.Config{
		AppConfig: cfg,
		Logger:    log,
		Database:  db,
	}

	diContainer, err := container.New(containerConfig)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize DI container: %v", err))
	}

	// Iniciar componentes do container
	if err := diContainer.Start(ctx); err != nil {
		log.Fatal(fmt.Sprintf("Failed to start container components: %v", err))
	}

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      diContainer.Handler(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Canal para capturar sinais do sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Canal para erros da aplicação
	errChan := make(chan error, 1)

	// Iniciar servidor HTTP em goroutine
	go func() {
		log.InfoWithFields("Starting HTTP server", map[string]interface{}{
			"address": server.Addr,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Iniciar reconexão automática em goroutine separada
	go connectOnStartup(diContainer, log)

	// Aguardar sinal de parada ou erro
	select {
	case sig := <-sigChan:
		log.InfoWithFields("Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})
	case err := <-errChan:
		log.ErrorWithFields("Application error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Graceful shutdown
	log.Info("Initiating graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Parar servidor HTTP
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.ErrorWithFields("Error shutting down HTTP server", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Parar componentes do container
	if err := diContainer.Stop(shutdownCtx); err != nil {
		log.ErrorWithFields("Error stopping container components", map[string]interface{}{
			"error": err.Error(),
		})
	}

	log.Info("Application shutdown completed successfully")
}

// connectOnStartup reconnects existing sessions automatically on startup
func connectOnStartup(container *container.Container, logger *logger.Logger) {
	const (
		startupDelay     = 3 * time.Second
		operationTimeout = 60 * time.Second
		sessionLimit     = 100
		reconnectDelay   = 1 * time.Second
	)

	time.Sleep(startupDelay)
	logger.Info("Starting automatic reconnection of existing sessions")

	sessionService := container.GetSessionService()
	if sessionService == nil {
		logger.Error("SessionService not available, skipping auto-connect")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Buscar sessões existentes que têm deviceJID (credenciais)
	sessions := getExistingSessions(ctx, sessionService, sessionLimit, logger)
	if len(sessions) == 0 {
		logger.Info("No existing sessions with credentials found, skipping auto-connect")
		return
	}

	logger.InfoWithFields("Starting auto-reconnect", map[string]interface{}{
		"total_sessions": len(sessions),
	})

	stats := reconnectSessions(ctx, sessions, sessionService, logger, reconnectDelay)

	logger.InfoWithFields("Auto-reconnect completed", map[string]interface{}{
		"connected": stats.connected,
		"skipped":   stats.skipped,
		"failed":    stats.failed,
	})
}

// getExistingSessions returns sessions that have saved credentials
func getExistingSessions(ctx context.Context, sessionService *services.SessionService, limit int, logger *logger.Logger) []sessionInfo {
	req := &contracts.ListSessionsRequest{
		Limit:  limit,
		Offset: 0,
	}

	response, err := sessionService.ListSessions(ctx, req)
	if err != nil {
		logger.ErrorWithFields("Failed to list sessions", map[string]interface{}{
			"error": err.Error(),
		})
		return nil
	}

	var sessionsWithCredentials []sessionInfo
	for _, sessionResponse := range response.Sessions {
		session := sessionResponse.Session
		if session.DeviceJID != "" {
			sessionsWithCredentials = append(sessionsWithCredentials, sessionInfo{
				ID:        session.ID,
				Name:      session.Name,
				DeviceJID: session.DeviceJID,
			})
		}
	}

	if len(sessionsWithCredentials) > 0 {
		logger.InfoWithFields("Found sessions with credentials", map[string]interface{}{
			"sessions_with_creds": len(sessionsWithCredentials),
		})
	}

	return sessionsWithCredentials
}

// reconnectSessions attempts to reconnect existing sessions
func reconnectSessions(ctx context.Context, sessions []sessionInfo, sessionService *services.SessionService, logger *logger.Logger, delay time.Duration) reconnectStats {
	stats := reconnectStats{}

	for _, session := range sessions {
		select {
		case <-ctx.Done():
			logger.Warn("Auto-reconnect cancelled due to timeout")
			return stats
		default:
		}

		result, err := sessionService.ConnectSession(ctx, session.ID)
		if err != nil {
			logger.ErrorWithFields("Failed to reconnect session", map[string]interface{}{
				"session_name": session.Name,
				"error":        err.Error(),
			})
			stats.failed++
		} else if result.Success {
			if result.QRCode != "" {
				stats.skipped++
			} else {
				logger.InfoWithFields("Session reconnected successfully", map[string]interface{}{
					"session_name": session.Name,
				})
				stats.connected++
			}
		} else {
			stats.failed++
		}

		if delay > 0 {
			time.Sleep(delay)
		}
	}

	return stats
}

type sessionInfo struct {
	ID        string
	Name      string
	DeviceJID string
}

type reconnectStats struct {
	connected int
	skipped   int
	failed    int
}

// runMigrations executa as migrações do banco de dados
func runMigrations(db *database.Database, log *logger.Logger) error {
	migrator := database.NewMigrator(db, log)
	return migrator.RunMigrations()
}

// printBanner exibe o banner da aplicação
func printBanner(cfg *config.Config) {
	fmt.Printf(`
    ███████╗██████╗ ██╗    ██╗ ██████╗  ██████╗ ████████╗
    ╚══███╔╝██╔══██╗██║    ██║██╔═══██╗██╔═══██╗╚══██╔══╝
      ███╔╝ ██████╔╝██║ █╗ ██║██║   ██║██║   ██║   ██║
     ███╔╝  ██╔═══╝ ██║███╗██║██║   ██║██║   ██║   ██║
    ███████╗██║     ╚███╔███╔╝╚██████╔╝╚██████╔╝   ██║
    ╚══════╝╚═╝      ╚══╝╚══╝  ╚═════╝  ╚═════╝    ╚═╝

    💬 WhatsApp API Gateway
    🚀 Version: %s | Environment: %s | Port: %d

`, appVersion, cfg.Environment, cfg.Server.Port)
}
