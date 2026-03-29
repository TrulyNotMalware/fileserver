package internal

import (
	"database/sql"
	"fileServer/configs"
	"fileServer/internal/auth"
	"fileServer/internal/db"
	"fileServer/internal/file"
	"fileServer/internal/logger"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	host       string
	port       string
	staticDir  string
	mux        *http.ServeMux
	jwtManager *auth.JWTManager
	cfg        *configs.Config
}

func NewServer(cfg *configs.Config) (*Server, error) {
	if cfg.PrivateKeyPath == "" {
		return nil, fmt.Errorf("auth.private_key_path is required")
	}
	if cfg.AdminPassword == "" || cfg.GuestPassword == "" {
		return nil, fmt.Errorf("auth.admin_password and auth.guest_password are required")
	}

	if err := prepareStaticDir(cfg.StaticDir, cfg.Permission, string(cfg.Mode)); err != nil {
		return nil, err
	}

	jwtManager, err := auth.NewJWTManager(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	database, err := db.New("./fileserver.db")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := seedUsers(database, cfg.AdminPassword, cfg.GuestPassword); err != nil {
		return nil, fmt.Errorf("failed to seed users: %w", err)
	}

	s := &Server{
		host:       cfg.Host,
		port:       cfg.Port,
		staticDir:  cfg.StaticDir,
		mux:        http.NewServeMux(),
		jwtManager: jwtManager,
		cfg:        cfg,
	}

	authHandler := auth.NewHandler(database, jwtManager, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	fileHandler := file.NewHandler(cfg.StaticDir)
	jwtMiddleware := auth.JWTMiddleware(jwtManager)
	InitRouter(s, authHandler, fileHandler, jwtMiddleware)

	fs := http.FileServer(http.Dir(s.staticDir))
	s.mux.Handle("GET /", jwtMiddleware(fs))

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

func seedUsers(database *sql.DB, adminPassword, guestPassword string) error {
	adminHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	guestHash, err := bcrypt.GenerateFromPassword([]byte(guestPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := db.UpsertUser(database, "admin", string(adminHash), db.RoleAdmin); err != nil {
		return err
	}
	return db.UpsertUser(database, "guest", string(guestHash), db.RoleGuest)
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	logger.Infof("file server listening on http://%s", addr)
	return http.ListenAndServe(addr, s.mux)
}
