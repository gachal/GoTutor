package server

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// requestLogger logs one line per request to the standard logger.
// Deliberately minimal — the desktop app's logs go to a file the user
// can find via Electron's app.getPath('logs'), so verbose JSON logs
// would just bloat that file without helping anyone debug.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %d %dms",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start).Milliseconds(),
		)
	}
}
