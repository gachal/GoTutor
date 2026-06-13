package server

import (
	"github.com/gin-gonic/gin"

	"gotutor/backend/internal/api"
)

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
	r.POST("/api/chapters/:id/submit", func(c *gin.Context) {
		api.HandleSubmit(c, db, s.cfg.GoBinary)
	})
	r.POST("/api/reset", func(c *gin.Context) {
		api.HandleReset(c, db)
	})
}
