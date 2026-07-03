package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

// sanitizeFilename rejects path traversal attempts
func sanitizeFilename(name string) error {
	if name == "" {
		return fmt.Errorf("filename is empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("filename must not contain path separators")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("filename must not contain '..'")
	}
	return nil
}

// resolvePath joins confdir + filename, then checks the result is still inside confdir
func resolvePath(confdir, filename string) (string, error) {
	if err := sanitizeFilename(filename); err != nil {
		return "", err
	}
	absFilesDir, _ := filepath.Abs(confdir)
	target := filepath.Join(absFilesDir, filename)
	absTarget, _ := filepath.Abs(target)
	if !strings.HasPrefix(absTarget, absFilesDir+string(os.PathSeparator)) && absTarget != absFilesDir {
		return "", fmt.Errorf("path traversal detected")
	}
	return absTarget, nil
}

func listFiles(confdir string) ([]FileInfo, error) {
	entries, err := os.ReadDir(confdir)
	if err != nil {
		return nil, fmt.Errorf("read confdir: %w", err)
	}
	var files []FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:     e.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
		})
	}
	return files, nil
}

func readFile(confdir, filename string) (string, error) {
	path, err := resolvePath(confdir, filename)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

func writeFile(confdir, filename, content string) error {
	path, err := resolvePath(confdir, filename)
	if err != nil {
		return err
	}
	// Ensure confdir exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	// Write to temp file then rename for atomicity
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

func deleteFile(confdir, filename string) error {
	path, err := resolvePath(confdir, filename)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

const lastCoreFile = ".last-core"

// readLastCore returns the name stored in .last-core, or "" if absent/empty.
func readLastCore() string {
	data, err := os.ReadFile(lastCoreFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// writeLastCore stores the core name in .last-core via atomic write.
func writeLastCore(name string) error {
	tmpPath := lastCoreFile + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(name), 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, lastCoreFile)
}

// clearLastCore removes .last-core.
func clearLastCore() {
	os.Remove(lastCoreFile)
}
