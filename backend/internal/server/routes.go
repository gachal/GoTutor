package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gotutor/backend/internal/api"
)

// corsMiddleware allows cross-origin requests from the packaged
// renderer (which runs on a file:// origin) to the localhost backend.
//
// We allow `*' because the backend binds to 127.0.0.1 only — remote
// hosts can't reach it regardless. Without these headers, the
// renderer's fetches are blocked by the browser's same-origin policy.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Accept-Language")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RegisterRoutes wires every API route onto the engine. Called once from
// Server.New. Health is wired here; chapter list/template/hint/reset
// land in Phase 2; submit lands in Phase 3.
//
// api handlers take *sql.DB directly so routes.go closes over s.DB()
// for each. We use inline closures instead of gin.HandlerFunc adapters
// to keep the call sites readable.
func (s *Server) RegisterRoutes(r *gin.Engine) {
	db := s.DB()

	r.GET("/api/health", s.handleHealth)

	r.GET("/api/chapters", func(c *gin.Context) {
		api.HandleListChapters(c, db)
	})
	r.GET("/api/chapters/:id/template", func(c *gin.Context) {
		api.HandleGetTemplate(c, db)
	})
	r.GET("/api/chapters/:id/hint", func(c *gin.Context) {
		api.HandleGetHint(c, db)
	})
	r.GET("/api/chapters/:id/solution", func(c *gin.Context) {
		api.HandleGetSolution(c, db)
	})
	r.POST("/api/chapters/:id/submit", func(c *gin.Context) {
		api.HandleSubmit(c, db, s.cfg.GoBinary)
	})
	r.POST("/api/reset", func(c *gin.Context) {
		api.HandleReset(c, db)
	})
}
