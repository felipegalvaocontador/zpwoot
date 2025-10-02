package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zpwoot/platform/config"
	"zpwoot/platform/container"
	"zpwoot/platform/database"
	"zpwoot/platform/logger"
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
