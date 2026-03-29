package file

import "time"

// Response DTOs

type entryResponse struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size,omitempty"` // dir is 0
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

type listResponse struct {
	Path    string          `json:"path"`
	Entries []entryResponse `json:"entries"`
}
