package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type APIError struct {
	Error string `json:"error"`
}

type APIResponse struct {
	OK    bool        `json:"ok,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIResponse{Error: msg})
}

func writeOK(w http.ResponseWriter) {
	writeJSON(w, 200, APIResponse{OK: true})
}

func writeData(w http.ResponseWriter, data interface{}) {
	writeJSON(w, 200, APIResponse{Data: data})
}

// --- File handlers ---

func handleListFiles(confdir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := listFiles(confdir)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if files == nil {
			files = []FileInfo{}
		}
		writeData(w, files)
	}
}

func handleReadFile(confdir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		content, err := readFile(confdir, filename)
		if err != nil {
			if isNotExist(err) {
				writeError(w, 404, fmt.Sprintf("file not found: %s", filename))
			} else {
				writeError(w, 400, err.Error())
			}
			return
		}
		writeData(w, map[string]string{"content": content})
	}
}

func handleWriteFile(confdir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")

		body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10MB limit
		if err != nil {
			writeError(w, 400, "failed to read request body")
			return
		}

		var req struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, 400, "invalid JSON: must be {\"content\": \"...\"}")
			return
		}

		if err := writeFile(confdir, filename, req.Content); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeOK(w)
	}
}

func handleDeleteFile(confdir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		if err := deleteFile(confdir, filename); err != nil {
			if isNotExist(err) {
				writeError(w, 404, fmt.Sprintf("file not found: %s", filename))
			} else {
				writeError(w, 500, err.Error())
			}
			return
		}
		writeOK(w)
	}
}

// --- Process handlers ---

func handleStatus(pm *ProcessManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeData(w, pm.Status())
	}
}

func handleStart(pm *ProcessManager, cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pm.Start(cfg.CoreStart); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeData(w, pm.Status())
	}
}

func handleStop(pm *ProcessManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pm.Stop(); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeOK(w)
	}
}

func handleTest(pm *ProcessManager, cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valid, errMsg := pm.TestConfig(cfg.CoreTest)
		writeData(w, map[string]interface{}{
			"valid": valid,
			"error": errMsg,
		})
	}
}

// --- Helpers ---

func isNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
