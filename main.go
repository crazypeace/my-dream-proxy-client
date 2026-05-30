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
)

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

	// Ensure confdir exists
	if err := os.MkdirAll(cfg.FilesDir, 0755); err != nil {
		log.Fatalf("failed to create confdir: %v", err)
	}

	// Log resolved paths
	absFilesDir, _ := filepath.Abs(cfg.FilesDir)
	log.Printf("files-dir: %s", absFilesDir)
	log.Printf("core-start: %s", cfg.CoreStart)
	log.Printf("core-test:  %s", cfg.CoreTest)

	// Create process manager
	pm := NewProcessManager()

	// Clean up core process on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("shutting down, stopping core...")
		pm.Cleanup()
		os.Exit(0)
	}()

	// Setup routes
	mux := http.NewServeMux()

	// File operations
	mux.HandleFunc("GET /api/files", handleListFiles(cfg.FilesDir))
	mux.HandleFunc("GET /api/files/{filename}", handleReadFile(cfg.FilesDir))
	mux.HandleFunc("PUT /api/files/{filename}", handleWriteFile(cfg.FilesDir))
	mux.HandleFunc("DELETE /api/files/{filename}", handleDeleteFile(cfg.FilesDir))

	// Process operations
	mux.HandleFunc("GET /api/core/status", handleStatus(pm))
	mux.HandleFunc("POST /api/core/start", handleStart(pm, cfg))
	mux.HandleFunc("POST /api/core/stop", handleStop(pm))
	mux.HandleFunc("POST /api/core/test", handleTest(pm, cfg))

	// Start server
	addr := net.JoinHostPort(cfg.Bind, cfg.Port)
	log.Printf("listening on %s", addr)
	fmt.Printf("my-dream-proxy-client listening on %s\n", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
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
