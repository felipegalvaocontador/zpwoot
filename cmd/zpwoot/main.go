package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

const (
	appName    = "zpwoot"
	appVersion = "2.0.0"
)

func main() {
	// Exibir banner
	printBanner()

	// Criar contexto principal da aplicação
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger
	log := logger.NewFromAppConfig(cfg)
	log.InfoWithFields("Starting zpwoot application", map[string]interface{}{
		"version":     appVersion,
		"environment": cfg.Environment,
		"port":        cfg.Server.Port,
	})

	// Inicializar aplicação
	app, err := initializeApplication(cfg, log)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize application: %v", err))
	}

	// Canal para capturar sinais do sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Canal para erros da aplicação
	errChan := make(chan error, 1)

	// Iniciar aplicação em goroutine
	go func() {
		log.InfoWithFields("Starting HTTP server", map[string]interface{}{
			"address": cfg.GetServerAddress(),
		})

		if err := app.Start(ctx); err != nil {
			errChan <- fmt.Errorf("server error: %w", err)
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

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.ErrorWithFields("Error during shutdown", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	log.Info("Application shutdown completed successfully")
}

// Application representa a aplicação principal
type Application struct {
	config *config.Config
	logger *logger.Logger
	// TODO: Adicionar dependências conforme implementação
	// database *database.Database
	// server   *server.Server
}

// initializeApplication inicializa todos os componentes da aplicação
func initializeApplication(cfg *config.Config, log *logger.Logger) (*Application, error) {
	log.Info("Initializing application components...")

	app := &Application{
		config: cfg,
		logger: log,
	}

	// TODO: Inicializar banco de dados
	// log.Info("Initializing database connection...")
	// db, err := database.NewFromAppConfig(cfg, log)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to initialize database: %w", err)
	// }
	// app.database = db

	// TODO: Executar migrações se habilitado
	// if cfg.Database.AutoMigrate {
	//     log.Info("Running database migrations...")
	//     if err := runMigrations(db, cfg, log); err != nil {
	//         return nil, fmt.Errorf("failed to run migrations: %w", err)
	//     }
	// }

	// TODO: Inicializar repositórios
	// log.Info("Initializing repositories...")

	// TODO: Inicializar serviços do core
	// log.Info("Initializing core services...")

	// TODO: Inicializar adapters
	// log.Info("Initializing adapters...")

	// TODO: Inicializar servidor HTTP
	// log.Info("Initializing HTTP server...")
	// server := server.New(cfg, log)
	// app.server = server

	log.Info("Application components initialized successfully")
	return app, nil
}

// Start inicia a aplicação
func (a *Application) Start(ctx context.Context) error {
	a.logger.Info("Starting application services...")

	// TODO: Iniciar serviços em background
	// go a.startBackgroundServices(ctx)

	// TODO: Iniciar servidor HTTP
	// return a.server.Start(ctx)

	// Placeholder - remover quando servidor for implementado
	a.logger.Info("Application started successfully (placeholder mode)")
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown para a aplicação graciosamente
func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down application...")

	// TODO: Parar servidor HTTP
	// if a.server != nil {
	//     a.logger.Info("Stopping HTTP server...")
	//     if err := a.server.Shutdown(ctx); err != nil {
	//         a.logger.ErrorWithFields("Error stopping HTTP server", map[string]interface{}{
	//             "error": err.Error(),
	//         })
	//     }
	// }

	// TODO: Parar serviços em background
	// a.logger.Info("Stopping background services...")

	// TODO: Fechar conexão com banco de dados
	// if a.database != nil {
	//     a.logger.Info("Closing database connection...")
	//     if err := a.database.Close(); err != nil {
	//         a.logger.ErrorWithFields("Error closing database", map[string]interface{}{
	//             "error": err.Error(),
	//         })
	//     }
	// }

	a.logger.Info("Application shutdown completed")
	return nil
}

// startBackgroundServices inicia serviços em background
func (a *Application) startBackgroundServices(ctx context.Context) {
	a.logger.Info("Starting background services...")

	// TODO: Implementar serviços em background
	// - Health check periodic
	// - Session monitoring
	// - Webhook retry mechanism
	// - Cleanup tasks

	<-ctx.Done()
	a.logger.Info("Background services stopped")
}

// runMigrations executa migrações do banco de dados
// func runMigrations(db *database.Database, cfg *config.Config, log *logger.Logger) error {
//     log.Info("Running database migrations...")
//
//     migrator := migrations.New(db, cfg.Database.MigrationsPath, log)
//     if err := migrator.Up(); err != nil {
//         return fmt.Errorf("failed to run migrations: %w", err)
//     }
//
//     log.Info("Database migrations completed successfully")
//     return nil
// }

// printBanner exibe banner da aplicação
func printBanner() {
	banner := `
 ███████╗██████╗ ██╗    ██╗ ██████╗  ██████╗ ████████╗
 ╚══███╔╝██╔══██╗██║    ██║██╔═══██╗██╔═══██╗╚══██╔══╝
   ███╔╝ ██████╔╝██║ █╗ ██║██║   ██║██║   ██║   ██║
  ███╔╝  ██╔═══╝ ██║███╗██║██║   ██║██║   ██║   ██║
 ███████╗██║     ╚███╔███╔╝╚██████╔╝╚██████╔╝   ██║
 ╚══════╝╚═╝      ╚══╝╚══╝  ╚═════╝  ╚═════╝    ╚═╝

 WhatsApp Business API Gateway - Clean Architecture
 Version: %s
 Environment: %s
`
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}
	fmt.Printf(banner, appVersion, env)
}