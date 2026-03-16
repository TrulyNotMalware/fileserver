package internal

import (
	"fmt"
	"net/http"
	"os"
)

type Server struct {
	host      string
	port      string
	staticDir string
	mux       *http.ServeMux
}

func NewServer(host, port, staticDir string, perm os.FileMode, mode string) (*Server, error) {
	if err := prepareStaticDir(staticDir, perm, mode); err != nil {
		return nil, err
	}

	s := &Server{
		host:      host,
		port:      port,
		staticDir: staticDir,
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func prepareStaticDir(dir string, perm os.FileMode, mode string) error {
	switch mode {
	case "create":
		return os.MkdirAll(dir, perm)
	case "error":
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("static dir %q does not exist", dir)
		}
	}
	return nil
}

func (s *Server) routes() {
	fs := http.FileServer(http.Dir(s.staticDir))
	s.mux.Handle("GET /", fs)
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	Infof("file server listening on http://%s", addr)
	return http.ListenAndServe(addr, s.mux)
}
