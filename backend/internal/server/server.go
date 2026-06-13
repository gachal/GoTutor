// Package server wires the Gin engine, database, and handlers into a
// single Server value with explicit Start/Shutdown lifecycle.
package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"

	"gotutor/backend/internal/config"
)

// Server holds all backend dependencies. Construct with New(), drive with
// Start/Shutdown. Handlers (Phase 2+) reach the DB via DB() and the
// chapter registry via the chapters package directly.
type Server struct {
	cfg     config.Config
	db      *sql.DB
	engine  *gin.Engine
	httpSrv *http.Server
}

// DB exposes the underlying *sql.DB so api handlers can run queries
// against the progress table. Returned pointer is shared; callers must
// not close it.
func (s *Server) DB() *sql.DB { return s.db }

// New constructs a Server. The caller passes an already-open *sql.DB
// (from db.Open); the server does not own its lifecycle (the caller
// defers Close).
func New(cfg config.Config, db *sql.DB) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())

	s := &Server{
		cfg:    cfg,
		db:     db,
		engine: r,
	}
	s.RegisterRoutes(r)
	return s
}

// Start binds the HTTP listener (scanning cfg.Port..cfg.PortMax if needed),
// writes the chosen port to cfg.PortFile, and serves until ctx is cancelled.
// Returns when Shutdown completes or the listener errors.
func (s *Server) Start(ctx context.Context) error {
	listener, port, err := s.bind()
	if err != nil {
		return err
	}

	if s.cfg.PortFile != "" {
		if err := writePortFile(s.cfg.PortFile, port); err != nil {
			return fmt.Errorf("write port file: %w", err)
		}
	}

	s.httpSrv = &http.Server{
		Handler:           s.engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := s.httpSrv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

// bind scans cfg.Port..cfg.PortMax for a free TCP port and returns a
// net.Listener bound to it. If cfg.Port is 0 the scan starts at 8081.
func (s *Server) bind() (net.Listener, int, error) {
	lo, hi := s.cfg.Port, s.cfg.PortMax
	if lo == 0 {
		lo = 8081
	}
	if hi < lo {
		hi = lo
	}
	var lastErr error
	for p := lo; p <= hi; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			return ln, p, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no ports available")
	}
	return nil, 0, fmt.Errorf("bind 127.0.0.1:%d..%d: %w", lo, hi, lastErr)
}

// Shutdown gracefully stops the HTTP server with a 5s deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(shutdownCtx)
}

// handleHealth reports backend liveness plus Go toolchain availability
// (the verifier needs `go` on PATH). The frontend polls this on startup
// to decide whether to show the install-Go screen.
func (s *Server) handleHealth(c *gin.Context) {
	goFound := false
	goVer := ""
	goPath := s.cfg.GoBinary
	if goPath == "" {
		goPath = "go"
	}
	if path, err := exec.LookPath(goPath); err == nil && path != "" {
		goFound = true
		if out, err := exec.Command(path, "version").Output(); err == nil {
			goVer = string(out)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"port":      s.cfg.Port,
		"goFound":   goFound,
		"goVersion": goVer,
	})
}

// writePortFile atomically writes the chosen port so Electron can poll it.
// We write to a temp file in the same dir then rename for atomicity.
func writePortFile(path string, port int) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(fmt.Sprintf("%d\n", port)), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
