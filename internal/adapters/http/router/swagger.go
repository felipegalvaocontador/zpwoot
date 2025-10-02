package router

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// setupSwaggerRoutes configura as rotas de documentação Swagger
func setupSwaggerRoutes(r *chi.Mux) {
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}