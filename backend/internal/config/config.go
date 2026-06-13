package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds all runtime configuration for the GoTutor backend.
// Defaults are derived from the OS (user config dir, temp dir) and can be
// overridden via flags or environment variables.
type Config struct {
	// HTTP port. If 0, the server scans Port..PortMax for a free port.
	Port     int
	PortMax  int
	PortFile string // path written with the chosen port so Electron can discover it

	// SQLite database path. Created if missing.
	DBPath string

	// TempDirRoot is the parent for verifier tempdirs (default: OS temp).
	TempDirRoot string

	// GoBinary path. Empty means resolve via PATH lookup.
	GoBinary string

	// VerifierOuterTimeout caps the whole submit handler.
	// GoTestTimeout caps the inner `go test -timeout` flag.
	// WaitDelay is the grace period before SIGKILL on cancel.
	VerifierOuterTimeout string
	GoTestTimeout        string
	WaitDelay            string
}

// Default returns the platform-aware default config. Callers then apply
// flag overrides in main.
func Default() Config {
	cfg := Config{
		Port:                 8081,
		PortMax:              8090,
		DBPath:               filepath.Join(userConfigDir(), "gotutor", "progress.db"),
		TempDirRoot:          filepath.Join(os.TempDir(), "gotutor"),
		GoBinary:             "",
		VerifierOuterTimeout: "15s",
		GoTestTimeout:        "10s",
		WaitDelay:            "2s",
	}
	if v := os.Getenv("GOTUTOR_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Port)
	}
	if v := os.Getenv("GOTUTOR_DB"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("GOTUTOR_GO_BINARY"); v != "" {
		cfg.GoBinary = v
	}
	return cfg
}

// RegisterFlags wires Config fields to flag.* pointers. main.go calls
// flag.Parse after wiring this.
func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.IntVar(&c.Port, "port", c.Port, "HTTP port (0 = scan Port..PortMax)")
	fs.IntVar(&c.PortMax, "port-max", c.PortMax, "upper bound for port scan")
	fs.StringVar(&c.PortFile, "portfile", c.PortFile, "path written with the chosen port for Electron discovery")
	fs.StringVar(&c.DBPath, "db", c.DBPath, "SQLite database path")
	fs.StringVar(&c.TempDirRoot, "tempdir", c.TempDirRoot, "parent directory for verifier tempdirs")
	fs.StringVar(&c.GoBinary, "go-binary", c.GoBinary, "Go toolchain binary path (empty = PATH lookup)")
}

// userConfigDir wraps os.UserConfigDir to allow test overrides.
func userConfigDir() string {
	d, err := os.UserConfigDir()
	if err != nil || d == "" {
		if home, err := os.UserHomeDir(); err == nil {
			if runtime.GOOS == "darwin" {
				return filepath.Join(home, "Library", "Application Support")
			}
			return filepath.Join(home, ".config")
		}
		return "."
	}
	return d
}
