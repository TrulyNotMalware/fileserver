package file

import (
	"encoding/json"
	"errors"
	"fileServer/internal/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	staticDir string
}

func NewHandler(staticDir string) *Handler {
	return &Handler{staticDir: staticDir}
}

// List godoc
// GET /files?path=subdir
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	logger.Debugf("List: path=%q", rel)

	absPath, err := h.safeJoin(rel)
	if err != nil {
		logger.Debugf("List: path traversal detected: %q", rel)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		logger.Debugf("List: path not found: %s", absPath)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if !info.IsDir() {
		logger.Debugf("List: path is not a directory: %s", absPath)
		http.Error(w, "Not a directory", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		logger.Errorf("List: failed to read dir %s: %v", absPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := listResponse{
		Path:    rel,
		Entries: make([]entryResponse, 0, len(entries)),
	}
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			logger.Debugf("List: failed to get file info for %s: %v", e.Name(), err)
			continue
		}
		resp.Entries = append(resp.Entries, entryResponse{
			Name:    fi.Name(),
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
			IsDir:   fi.IsDir(),
		})
	}

	logger.Debugf("List: returned %d entries for path=%q", len(resp.Entries), rel)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	logger.Debugf("Download: path=%q", rel)

	if rel == "" {
		logger.Debugf("Download: path is required")
		http.Error(w, "Bad Request: path is required", http.StatusBadRequest)
		return
	}

	absPath, err := h.safeJoin(rel)
	if err != nil {
		logger.Debugf("Download: path traversal detected: %q", rel)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		logger.Debugf("Download: file not found: %s", absPath)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if info.IsDir() {
		logger.Debugf("Download: path is a directory: %s", absPath)
		http.Error(w, "Bad Request: path is a directory", http.StatusBadRequest)
		return
	}

	logger.Infof("Download: serving file %s (size: %d bytes)", absPath, info.Size())
	w.Header().Set("Content-Disposition", "attachment; filename=\""+info.Name()+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, absPath)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	const maxSize = 512 << 20 // 512 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		logger.Debugf("Upload: failed to parse multipart form: %v", err)
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	rel := r.URL.Query().Get("path")
	logger.Debugf("Upload: destination path=%q", rel)

	dirPath, err := h.safeJoin(rel)
	if err != nil {
		logger.Debugf("Upload: path traversal detected: %q", rel)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		logger.Errorf("Upload: failed to create dir %s: %v", dirPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logger.Debugf("Upload: file field missing: %v", err)
		http.Error(w, "Bad Request: file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	logger.Debugf("Upload: receiving file %q (%d bytes)", header.Filename, header.Size)

	destPath, err := h.safeJoin(filepath.Join(rel, header.Filename))
	if err != nil {
		logger.Debugf("Upload: path traversal detected on filename %q", header.Filename)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	dst, err := os.Create(destPath)
	if err != nil {
		logger.Errorf("Upload: failed to create file %s: %v", destPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logger.Errorf("Upload: failed to write file %s: %v", destPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Infof("Upload: saved file %s (%d bytes)", destPath, header.Size)
	w.WriteHeader(http.StatusCreated)
}

// safeJoin is path traversal prevention function that returns absolute path within staticDir
func (h *Handler) safeJoin(rel string) (string, error) {
	absStatic, err := filepath.Abs(h.staticDir)
	if err != nil {
		return "", err
	}

	joined := filepath.Join(absStatic, filepath.FromSlash(rel))
	if !strings.HasPrefix(joined, absStatic+string(os.PathSeparator)) && joined != absStatic {
		return "", errors.New("path traversal detected")
	}

	return joined, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
