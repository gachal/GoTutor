package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// routeDef describes one API route plus its handler. Kept as a slice so
// routes.go reads as a flat table — easy to scan when adding endpoints.
type routeDef struct {
	method  string
	path    string
	handler gin.HandlerFunc
}

// RegisterRoutes wires every API route onto the engine. Called once from
// Server.Start. Health is wired here so /api/health works before any
// chapter/verifier handler is implemented.
func (s *Server) RegisterRoutes(r *gin.Engine) {
	routes := []routeDef{
		{http.MethodGet, "/api/health", s.handleHealth},
		// Phase 2 fills these in:
		{http.MethodGet, "/api/chapters", stubNotImplemented},
		{http.MethodGet, "/api/chapters/:id/template", stubNotImplemented},
		{http.MethodGet, "/api/chapters/:id/hint", stubNotImplemented},
		{http.MethodPost, "/api/chapters/:id/submit", stubNotImplemented},
		{http.MethodPost, "/api/reset", stubNotImplemented},
	}

	api := r.Group("")
	for _, rt := range routes {
		api.Handle(rt.method, rt.path, rt.handler)
	}
}

// stubNotImplemented returns 501 for endpoints wired in Phase 1 but
// implemented in later phases. Replaced chapter-by-chapter.
func stubNotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not yet implemented",
		"path":  c.Request.URL.Path,
	})
}
