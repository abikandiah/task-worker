package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Set up API routes
func (server *Server) setupRoutes() {
	server.router.Get("/health", server.handleHealth)

	server.router.Route("/api/v1", func(r chi.Router) {
		// Apply authentication to all routes
		r.Use(server.authenticateMiddleware)
		r.Use(server.contentTypeMiddleware)

		// Items endpoints
		r.Route("/job", func(r chi.Router) {
			r.Get("/", server.listItems)
			// r.Post("/", server.createItem)

			// r.Route("/{id}", func(r chi.Router) {
			// r.Get("/", server.getItem)
			// r.Put("/", server.updateItem)
			// r.Delete("/", server.deleteItem)
		})
	})

	// Easy to add more resource groups
	// r.Route("/users", func(r chi.Router) { ... })
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items := make([]struct{}, 0, 10)
	s.respondJSON(w, http.StatusOK, items)
}
