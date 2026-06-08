package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

// getCore extracts the core name from the URL path and returns its config and process manager.
func (app *App) getCore(r *http.Request) (*CoreConfig, *ProcessManager, error) {
	coreName := r.PathValue("core")
	core, ok := app.Config.Cores[coreName]
	if !ok {
		return nil, nil, fmt.Errorf("core not found: %s", coreName)
	}
	pm, ok := app.PMs[coreName]
	if !ok {
		return nil, nil, fmt.Errorf("process manager not found: %s", coreName)
	}
	return core, pm, nil
}

// --- File handlers ---

func handleListFiles(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		core, _, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		files, err := listFiles(core.FilesDir)
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

func handleReadFile(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		core, _, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		filename := r.PathValue("filename")
		content, err := readFile(core.FilesDir, filename)
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

func handleWriteFile(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		core, _, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
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

		if err := writeFile(core.FilesDir, filename, req.Content); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeOK(w)
	}
}

func handleDeleteFile(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		core, _, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		filename := r.PathValue("filename")
		if err := deleteFile(core.FilesDir, filename); err != nil {
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

func handleStatus(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, pm, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		writeData(w, pm.Status())
	}
}

func handleStart(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		coreName := r.PathValue("core")
		if err := app.StartExclusive(coreName); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		writeData(w, app.PMs[coreName].Status())
	}
}

func handleStop(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, pm, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		if err := pm.Stop(); err != nil {
			log.Printf("core/stop failed: %v", err)
			writeError(w, 500, err.Error())
			return
		}
		log.Println("core/stop succeeded")
		writeData(w, pm.Status())
	}
}

func handleTest(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		core, pm, err := app.getCore(r)
		if err != nil {
			writeError(w, 404, err.Error())
			return
		}
		valid, errMsg := pm.TestConfig(core.CoreTest)
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
