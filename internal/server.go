package internal

import (
	"database/sql"
	"fileServer/internal/auth"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	host            string
	port            string
	staticDir       string
	mux             *http.ServeMux
	db              *sql.DB
	jwtManager      *auth.JWTManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewServer(
	host, port, staticDir string,
	perm os.FileMode,
	mode string,
	database *sql.DB,
	jwtManager *auth.JWTManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) (*Server, error) {
	if err := prepareStaticDir(staticDir, perm, mode); err != nil {
		return nil, err
	}

	s := &Server{
		host:            host,
		port:            port,
		staticDir:       staticDir,
		mux:             http.NewServeMux(),
		db:              database,
		jwtManager:      jwtManager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
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
	authHandler := auth.NewHandler(s.db, s.jwtManager, s.accessTokenTTL, s.refreshTokenTTL)

	s.mux.HandleFunc("POST /auth/login", authHandler.Login)
	s.mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	s.mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	fs := http.FileServer(http.Dir(s.staticDir))
	s.mux.Handle("GET /", auth.JWTMiddleware(s.jwtManager)(fs))
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	Infof("file server listening on http://%s", addr)
	return http.ListenAndServe(addr, s.mux)
}
