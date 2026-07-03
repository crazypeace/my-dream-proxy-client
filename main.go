package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// App holds global state: config and per-core process managers.
type App struct {
	Config *Config
	PMs    map[string]*ProcessManager
}

func main() {
	cfg := LoadConfig()

	// Setup log output
	if cfg.Log != "" {
		f, err := os.OpenFile(cfg.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Ensure confdir exists for each core, log resolved paths
	for name, core := range cfg.Cores {
		if core.FilesDir != "" {
			if err := os.MkdirAll(core.FilesDir, 0755); err != nil {
				log.Fatalf("failed to create confdir for %s: %v", name, err)
			}
		}
		absFilesDir, _ := filepath.Abs(core.FilesDir)
		log.Printf("[%s] files-dir:  %s", name, absFilesDir)
		log.Printf("[%s] core-start: %s", name, core.CoreStart)
		log.Printf("[%s] core-test:  %s", name, core.CoreTest)
	}

	// Create process managers for each core
	app := &App{
		Config: cfg,
		PMs:    make(map[string]*ProcessManager),
	}
	for name := range cfg.Cores {
		app.PMs[name] = NewProcessManager()
	}

	// Clean up all core processes on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("shutting down, stopping all cores...")
		for name, pm := range app.PMs {
			log.Printf("stopping %s...", name)
			pm.Cleanup()
		}
		os.Exit(0)
	}()

	// Setup routes
	mux := http.NewServeMux()

	// File operations — /api/{core}/files
	mux.HandleFunc("GET /api/{core}/files", handleListFiles(app))
	mux.HandleFunc("GET /api/{core}/files/{filename}", handleReadFile(app))
	mux.HandleFunc("PUT /api/{core}/files/{filename}", handleWriteFile(app))
	mux.HandleFunc("DELETE /api/{core}/files/{filename}", handleDeleteFile(app))

	// Process operations — /api/{core}/core
	mux.HandleFunc("GET /api/{core}/core/status", handleStatus(app))
	mux.HandleFunc("POST /api/{core}/core/start", handleStart(app))
	mux.HandleFunc("POST /api/{core}/core/stop", handleStop(app))
	mux.HandleFunc("POST /api/{core}/core/test", handleTest(app))

	// Start server
	addr := net.JoinHostPort(cfg.Bind, cfg.Port)
	log.Printf("listening on %s", addr)
	fmt.Printf("my-dream-proxy-client listening on %s\n", addr)

	// Auto-start last core (non-blocking, after server is ready)
	go func() {
		lastCore := readLastCore()
		if lastCore == "" {
			return
		}
		if _, ok := cfg.Cores[lastCore]; !ok {
			log.Printf("[auto-start] skipped: core %q not in config", lastCore)
			return
		}
		log.Printf("[auto-start] starting core: %s", lastCore)
		if err := app.StartExclusive(lastCore); err != nil {
			log.Printf("[auto-start] failed: %v", err)
			return
		}
		// Confirm process is actually running after a brief settle
		time.Sleep(500 * time.Millisecond)
		if app.PMs[lastCore].Status().Running {
			log.Printf("[auto-start] confirmed running: %s (pid %d)", lastCore, app.PMs[lastCore].Status().PID)
		} else {
			log.Printf("[auto-start] process exited immediately: %s (binary missing?)", lastCore)
		}
	}()

	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// StartExclusive starts the named core, stopping any other running core first.
// Only one core may run at a time.
func (app *App) StartExclusive(coreName string) error {
	core, ok := app.Config.Cores[coreName]
	if !ok {
		return fmt.Errorf("core not found: %s", coreName)
	}
	pm, ok := app.PMs[coreName]
	if !ok {
		return fmt.Errorf("process manager not found: %s", coreName)
	}

	// Stop any other running core
	for name, otherPM := range app.PMs {
		if name == coreName {
			continue
		}
		if otherPM.Status().Running {
			log.Printf("stopping %s to start %s", name, coreName)
			otherPM.Stop()
		}
	}

	return pm.Start(core.CoreStart)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
