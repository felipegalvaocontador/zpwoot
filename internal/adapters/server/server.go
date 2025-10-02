package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"zpwoot/internal/adapters/server/router"
	"zpwoot/internal/services"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

type Server struct {
	config         *config.Config
	logger         *logger.Logger
	httpServer     *http.Server
	sessionService *services.SessionService
	messageService *services.MessageService
	groupService   *services.GroupService
}

type Config struct {
	Config         *config.Config
	Logger         *logger.Logger
	SessionService *services.SessionService
	MessageService *services.MessageService
	GroupService   *services.GroupService
}

func New(cfg *Config) *Server {
	return &Server{
		config:         cfg.Config,
		logger:         cfg.Logger,
		sessionService: cfg.SessionService,
		messageService: cfg.MessageService,
		groupService:   cfg.GroupService,
	}
}

func (s *Server) Start(ctx context.Context) error {

	handler := router.SetupRoutes(
		s.config,
		s.logger,
		s.sessionService,
		s.messageService,
		s.groupService,
	)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	s.logger.InfoWithFields("Starting HTTP server", map[string]interface{}{
		"port":          s.config.Server.Port,
		"read_timeout":  s.config.Server.ReadTimeout,
		"write_timeout": s.config.Server.WriteTimeout,
		"idle_timeout":  s.config.Server.IdleTimeout,
	})

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorWithFields("HTTP server failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	s.logger.InfoWithFields("HTTP server started successfully", map[string]interface{}{
		"address": s.httpServer.Addr,
	})

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info("Stopping HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorWithFields("Failed to shutdown HTTP server gracefully", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	s.logger.Info("HTTP server stopped successfully")
	return nil
}

func (s *Server) Handler() http.Handler {
	return router.SetupRoutes(
		s.config,
		s.logger,
		s.sessionService,
		s.messageService,
		s.groupService,
	)
}

func (s *Server) Address() string {
	if s.httpServer != nil {
		return s.httpServer.Addr
	}
	return fmt.Sprintf(":%d", s.config.Server.Port)
}

func (s *Server) IsRunning() bool {
	return s.httpServer != nil
}
