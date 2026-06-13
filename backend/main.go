// Command gotutor-backend is the HTTP sidecar spawned by the Electron
// main process. It owns the SQLite database, serves chapter templates
// and hints, and runs the verifier (Phase 3) on user code submissions.
//
// Lifecycle:
//   1. Parse flags + env (config.Config).
//   2. Open the SQLite database (db.Open), applying migrations.
//   3. Construct the Server, bind the HTTP port (scanning 8081..8090
//      on collision), write the chosen port to ~/.gotutor/backend.port.
//   4. Serve until SIGINT or SIGTERM, then Shutdown with a 5s deadline.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gotutor/backend/internal/config"
	"gotutor/backend/internal/db"
	"gotutor/backend/internal/server"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gotutor-backend:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg := config.Default()

	fs := flag.NewFlagSet("gotutor-backend", flag.ContinueOnError)
	// Default portfile lives in the user's home so Electron can find it
	// predictably across dev and packaged builds.
	if cfg.PortFile == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.PortFile = home + "/.gotutor/backend.port"
		}
	}
	cfg.RegisterFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	srv := server.New(cfg, database)

	// Cancel on SIGINT/SIGTERM so Server.Start unblocks and Shutdown runs.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("gotutor-backend starting (db=%s)", cfg.DBPath)
	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	log.Printf("gotutor-backend stopped cleanly")
	return nil
}
