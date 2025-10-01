package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"zpwoot/internal/adapters/http/handlers"
	httpMiddleware "zpwoot/internal/adapters/http/middleware"
	"zpwoot/internal/services"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

// Router configuração do roteador HTTP
type Router struct {
	config         *config.Config
	logger         *logger.Logger
	sessionService *services.SessionService
}

// NewRouter cria nova instância do router
func NewRouter(
	cfg *config.Config,
	logger *logger.Logger,
	sessionService *services.SessionService,
) *Router {
	return &Router{
		config:         cfg,
		logger:         logger,
		sessionService: sessionService,
	}
}

// SetupRoutes configura todas as rotas da aplicação
func (rt *Router) SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Configurar middlewares globais
	rt.setupGlobalMiddlewares(r)

	// Configurar rotas públicas
	rt.setupPublicRoutes(r)

	// Configurar rotas protegidas
	rt.setupProtectedRoutes(r)

	return r
}

// setupGlobalMiddlewares configura middlewares que se aplicam a todas as rotas
func (rt *Router) setupGlobalMiddlewares(r *chi.Mux) {
	// Middleware de recuperação de panic
	r.Use(middleware.Recoverer)

	// Middleware de timeout
	r.Use(middleware.Timeout(30 * time.Second))

	// Middleware de compressão
	r.Use(middleware.Compress(5))

	// Middleware de headers de segurança
	r.Use(httpMiddleware.SecurityHeaders())

	// Middleware de CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   rt.getAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300, // 5 minutos
	}))

	// Middleware de logging de erro (captura panics)
	r.Use(httpMiddleware.ErrorLogger(rt.logger))

	// Middleware de logging HTTP
	if rt.config.Log.Level == "debug" {
		// Em debug, usar logging detalhado
		r.Use(httpMiddleware.HTTPLogger(rt.logger))
	} else {
		// Em produção, usar logging normal
		r.Use(httpMiddleware.HTTPLogger(rt.logger))
	}

	// Middleware de performance (log requisições lentas)
	r.Use(httpMiddleware.PerformanceLogger(rt.logger, 2*time.Second))

	// Middleware de request ID
	r.Use(middleware.RequestID)

	// Middleware de real IP
	r.Use(middleware.RealIP)
}

// setupPublicRoutes configura rotas que não requerem autenticação
func (rt *Router) setupPublicRoutes(r *chi.Mux) {
	// Health check
	r.Get("/health", rt.handleHealth)

	// Swagger documentation
	if rt.config.App.Environment != "production" {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	// Webhook endpoints (não requerem API key)
	r.Route("/webhook", func(r chi.Router) {
		// TODO: Implementar webhooks
		r.Get("/events", rt.handleWebhookEvents)
	})
}

// setupProtectedRoutes configura rotas que requerem autenticação
func (rt *Router) setupProtectedRoutes(r *chi.Mux) {
	// Aplicar middleware de autenticação
	r.Use(httpMiddleware.APIKeyAuth(rt.config, rt.logger))

	// Rotas de sessões
	rt.setupSessionRoutes(r)

	// TODO: Adicionar outras rotas protegidas
	// rt.setupMessageRoutes(r)
	// rt.setupContactRoutes(r)
	// rt.setupGroupRoutes(r)
}

// setupSessionRoutes configura rotas relacionadas a sessões
func (rt *Router) setupSessionRoutes(r *chi.Mux) {
	sessionHandler := handlers.NewSessionHandler(rt.sessionService, rt.logger)

	r.Route("/sessions", func(r chi.Router) {
		// Operações globais de sessões
		r.Post("/create", sessionHandler.CreateSession)
		r.Get("/list", sessionHandler.ListSessions)
		r.Get("/stats", sessionHandler.GetSessionStats)

		// Operações específicas de sessão
		r.Route("/{sessionId}", func(r chi.Router) {
			// Informações da sessão
			r.Get("/info", sessionHandler.GetSessionInfo)
			r.Delete("/delete", sessionHandler.DeleteSession)

			// Controle de conexão
			r.Post("/connect", sessionHandler.ConnectSession)
			r.Post("/disconnect", sessionHandler.DisconnectSession)

			// QR Code
			r.Get("/qr", sessionHandler.GetQRCode)
			r.Post("/qr/generate", sessionHandler.GenerateQRCode)

			// Proxy
			r.Post("/proxy/set", sessionHandler.SetProxy)
			r.Get("/proxy", sessionHandler.GetProxy)

			// TODO: Adicionar rotas de mensagens, contatos, grupos, etc.
			// r.Route("/messages", rt.setupMessageSubRoutes)
			// r.Route("/contacts", rt.setupContactSubRoutes)
			// r.Route("/groups", rt.setupGroupSubRoutes)
		})
	})
}

// getAllowedOrigins retorna origens permitidas para CORS
func (rt *Router) getAllowedOrigins() []string {
	if rt.config.App.Environment == "development" {
		return []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		}
	}

	// Em produção, configurar origens específicas
	return []string{
		// TODO: Configurar origens de produção
		"https://yourdomain.com",
	}
}

// ===== HANDLERS BÁSICOS =====

// handleHealth handler para health check
func (rt *Router) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"zpwoot"}`))
}

// handleWebhookEvents handler para listar eventos de webhook suportados
func (rt *Router) handleWebhookEvents(w http.ResponseWriter, r *http.Request) {
	events := []string{
		"message.received",
		"message.sent",
		"session.connected",
		"session.disconnected",
		"qr.generated",
		"qr.scanned",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Resposta simples para eventos suportados
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"supported_events": events,
			"webhook_format":   "json",
			"delivery_method":  "http_post",
		},
		"message": "Supported webhook events retrieved successfully",
	}

	// Encoding manual para evitar dependência
	w.Write([]byte(`{
		"success": true,
		"data": {
			"supported_events": [
				"message.received",
				"message.sent", 
				"session.connected",
				"session.disconnected",
				"qr.generated",
				"qr.scanned"
			],
			"webhook_format": "json",
			"delivery_method": "http_post"
		},
		"message": "Supported webhook events retrieved successfully"
	}`))
}

// ===== MIDDLEWARE HELPERS =====

// LogRoutes registra todas as rotas configuradas (útil para debug)
func (rt *Router) LogRoutes(r *chi.Mux) {
	if rt.config.Log.Level != "debug" {
		return
	}

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		rt.logger.DebugWithFields("Route registered", map[string]interface{}{
			"method": method,
			"route":  route,
		})
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		rt.logger.ErrorWithFields("Failed to walk routes", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// ===== FUTURE ROUTE GROUPS =====

// TODO: Implementar quando os serviços estiverem prontos

// func (rt *Router) setupMessageRoutes(r *chi.Mux) {
//     messageHandler := handlers.NewMessageHandler(rt.messageService, rt.logger)
//     
//     r.Route("/messages", func(r chi.Router) {
//         r.Post("/send/text", messageHandler.SendText)
//         r.Post("/send/media", messageHandler.SendMedia)
//         // ... outros endpoints de mensagem
//     })
// }

// func (rt *Router) setupContactRoutes(r *chi.Mux) {
//     contactHandler := handlers.NewContactHandler(rt.contactService, rt.logger)
//     
//     r.Route("/contacts", func(r chi.Router) {
//         r.Get("/list", contactHandler.ListContacts)
//         r.Get("/{contactId}", contactHandler.GetContact)
//         // ... outros endpoints de contato
//     })
// }

// func (rt *Router) setupGroupRoutes(r *chi.Mux) {
//     groupHandler := handlers.NewGroupHandler(rt.groupService, rt.logger)
//     
//     r.Route("/groups", func(r chi.Router) {
//         r.Get("/list", groupHandler.ListGroups)
//         r.Post("/create", groupHandler.CreateGroup)
//         // ... outros endpoints de grupo
//     })
// }
